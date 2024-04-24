package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/kachick/selfup/internal/migrate"
	"github.com/kachick/selfup/internal/runner"
	"golang.org/x/term"
	"golang.org/x/xerrors"
)

var (
	// Used in goreleaser
	version = "dev"
	commit  = "none"

	revision = "rev"
)

type Result struct {
	Path       string
	FileResult runner.Result
}

func main() {
	versionFlag := flag.Bool("version", false, "print the version of this program")

	sharedFlags := flag.NewFlagSet("run|list", flag.ExitOnError)
	prefixFlag := sharedFlags.String("prefix", " selfup ", "prefix to begin json")
	skipByFlag := sharedFlags.String("skip-by", "", "skip to run if the line contains this string")
	checkFlag := sharedFlags.Bool("check", false, "exit as error if found changes")
	noColorFlag := sharedFlags.Bool("no-color", false, "disable color output")

	const usage = `Usage: selfup [SUB] [OPTIONS] [PATH]...

$ selfup run .github/workflows/*.yml
$ selfup run --prefix='# Update with this json: ' --skip-by='nix run' .github/workflows/*.yml
$ selfup list .github/workflows/*.yml
$ selfup migrate .github/workflows/have_beta_schema.yml
$ selfup --version
`

	flag.Usage = func() {
		// https://github.com/golang/go/issues/57059#issuecomment-1336036866
		fmt.Printf("%s", usage+"\n\n")
		fmt.Println("Usage of command:")
		flag.PrintDefaults()
		fmt.Println("")
		sharedFlags.Usage()
	}

	if len(commit) >= 7 {
		revision = commit[:7]
	}
	version := fmt.Sprintf("%s\n", "selfup"+" "+version+" "+"("+revision+")")

	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		return
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	subCommand := os.Args[1]
	isListMode := subCommand == "list"
	isRunMode := subCommand == "run"
	isMigrateMode := subCommand == "migrate"
	if isMigrateMode {
		paths := os.Args[2:]
		for _, path := range paths {
			isMigrated, err := migrate.Migrate(path)
			if err != nil {
				log.Fatalf("%+v", err)
			}
			if isMigrated {
				log.Println(path + ": migrated schema beta -> v1")
			}
		}

		return
	}

	if !(isListMode || isRunMode) {
		flag.Usage()
		log.Fatalf("Specified unexpected subcommand `%s`", subCommand)
	}

	sharedFlags.Parse(os.Args[2:])
	paths := sharedFlags.Args()
	prefix := *prefixFlag
	skipBy := *skipByFlag
	isCheckMode := *checkFlag
	isColor := term.IsTerminal(int(os.Stdout.Fd())) && !(*noColorFlag)

	if prefix == "" {
		flag.Usage()
		log.Fatalf("%+v", xerrors.New("No prefix is specified"))
	}

	wg := &sync.WaitGroup{}
	results := make(chan Result, len(paths))
	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			fileResult := func() runner.Result {
				file, err := os.Open(path)
				if err != nil {
					log.Fatalf("%s: %+v", path, err)
				}
				defer file.Close()

				fr, err := runner.DryRun(file, prefix, skipBy)
				if err != nil {
					log.Fatalf("%s: %+v", path, err)
				}
				return fr
			}()

			isDirty := fileResult.ChangedCount > 0

			if isRunMode && isDirty {
				err := os.WriteFile(path, []byte(strings.Join(fileResult.NewLines, "\n")+"\n"), os.ModePerm)
				if err != nil {
					log.Fatalf("%s: %+v", path, err)
				}
			}

			results <- Result{
				Path:       path,
				FileResult: fileResult,
			}
		}(path)
	}
	wg.Wait()
	close(results)
	total := 0
	changed := 0
	for r := range results {
		fr := r.FileResult
		total += fr.Total
		changed += fr.ChangedCount
		for _, t := range fr.Targets {
			estimation := " "
			suffix := ""
			replacer := t.Replacer
			if t.IsChanged {
				estimation = "✓"
				if isColor {
					green := color.New(color.FgGreen).SprintFunc()
					estimation = green(estimation)
					replacer = green(t.Replacer)
				}
				suffix = fmt.Sprintf(" => %s", replacer)
			}
			fmt.Printf("%s %s:%d: %s%s\n", estimation, r.Path, t.LineNumber, t.Extracted, suffix)
		}
	}
	fmt.Println()
	if isListMode {
		fmt.Printf("%d/%d items will be replaced\n", changed, total)
	} else {
		fmt.Printf("%d/%d items have been replaced\n", changed, total)
	}
	if isCheckMode && (changed > 0) {
		os.Exit(1)
	}
}
