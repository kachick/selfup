package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
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
)

type Result struct {
	Path       string
	FileResult runner.Result
	Err        error
}

func main() {
	versionFlag := flag.Bool("version", false, "print the version of this program")

	sharedFlags := flag.NewFlagSet("run|list", flag.ExitOnError)
	prefixFlag := sharedFlags.String("prefix", "\\s*[#;/]* selfup ", "start JSON after this pattern(RE2)")
	skipByFlag := sharedFlags.String("skip-by", "", "skip to run if the line contains this string")
	checkFlag := sharedFlags.Bool("check", false, "exit as error if found changes")
	noColorFlag := sharedFlags.Bool("no-color", false, "disable color output")

	const usage = `Usage: selfup [SUB] [OPTIONS] [PATH]...

$ selfup run .github/workflows/*.yml
$ selfup list --check .github/workflows/*.yml
`

	flag.Usage = func() {
		// https://github.com/golang/go/issues/57059#issuecomment-1336036866
		fmt.Printf("%s", usage+"\n\n")
		fmt.Println("Usage of command:")
		flag.PrintDefaults()
		fmt.Println("")
		sharedFlags.Usage()
	}

	version := fmt.Sprintf("%s\n", "selfup"+" "+version)

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
	prefixStr := *prefixFlag
	skipBy := *skipByFlag
	isCheckMode := *checkFlag
	isColor := term.IsTerminal(int(os.Stdout.Fd())) && !(*noColorFlag)

	if prefixStr == "" {
		flag.Usage()
		log.Fatalf("%+v", xerrors.New("No prefix is specified"))
	}

	prefix, err := regexp.Compile(prefixStr)
	if err != nil {
		log.Fatalf("Given an invalid regex: `%v`", err)
	}

	wg := &sync.WaitGroup{}
	results := make(chan Result, len(paths))
	for _, path := range paths {
		wg.Go(func() {
			fileResult, err := func() (runner.Result, error) {
				file, err := os.Open(path)
				if err != nil {
					return runner.Result{}, err
				}
				defer file.Close()

				return runner.DryRun(file, prefix, skipBy)
			}()

			if err != nil {
				results <- Result{
					Path: path,
					Err:  err,
				}
				return
			}

			isDirty := fileResult.ChangedCount > 0

			if isRunMode && isDirty {
				err := os.WriteFile(path, []byte(strings.Join(fileResult.NewLines, "\n")+"\n"), os.ModePerm)
				if err != nil {
					results <- Result{
						Path: path,
						Err:  err,
					}
					return
				}
			}

			results <- Result{
				Path:       path,
				FileResult: fileResult,
			}
		})
	}
	wg.Wait()
	close(results)
	total := 0
	changed := 0
	hasError := false
	for r := range results {
		if r.Err != nil {
			log.Printf("%s: %+v", r.Path, r.Err)
			hasError = true
			continue
		}
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

	if hasError || (isCheckMode && (changed > 0)) {
		os.Exit(1)
	}
}
