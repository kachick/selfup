package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	updater "github.com/kachick/selfup/internal"
	"golang.org/x/term"
	"golang.org/x/xerrors"
)

var (
	// Used in goreleaser
	version = "dev"
	commit  = "none"

	revision = "rev"
)

func main() {
	versionFlag := flag.Bool("version", false, "print the version of this program")

	sharedFlags := flag.NewFlagSet("run|list", flag.ExitOnError)
	prefixFlag := sharedFlags.String("prefix", " selfup ", "prefix to begin json")
	skipByFlag := sharedFlags.String("skip-by", "", "skip to run if the line contains this string")
	noColorFlag := sharedFlags.Bool("no-color", false, "disable color output")

	const usage = `Usage: selfup [SUB] [OPTIONS] [PATH]...

$ selfup run .github/workflows/*.yml
$ selfup run --prefix='# Update with this json: ' --skip-by='nix run' .github/workflows/*.yml
$ selfup list .github/workflows/*.yml
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

	if !(isListMode || isRunMode) {
		flag.Usage()
		log.Fatalf("Specified unexpected subcommand `%s`", subCommand)
	}

	sharedFlags.Parse(os.Args[2:])
	paths := sharedFlags.Args()
	prefix := *prefixFlag
	skipBy := *skipByFlag
	isColor := term.IsTerminal(int(os.Stdout.Fd())) && !(*noColorFlag)

	if prefix == "" {
		flag.Usage()
		log.Fatalf("%+v", xerrors.New("No prefix is specified"))
	}

	wg := &sync.WaitGroup{}
	results := make(chan updater.Result, len(paths))
	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			result, err := updater.Update(path, prefix, skipBy, isColor)
			if err != nil {
				log.Fatalf("%+v", err)
			}
			results <- result
			isDirty := result.Changed > 0

			if isRunMode && isDirty {
				err := os.WriteFile(path, []byte(strings.Join(result.Lines, "\n")+"\n"), os.ModePerm)
				if err != nil {
					log.Fatalf("%+v", xerrors.Errorf("%s: %w", path, err))
				}
			}
		}(path)
	}
	wg.Wait()
	close(results)
	total := 0
	changed := 0
	for r := range results {
		total += r.Total
		changed += r.Changed
	}
	fmt.Println()
	if isListMode {
		fmt.Printf("%d/%d items will be replaced\n", changed, total)
	} else {
		fmt.Printf("%d/%d items have been replaced\n", changed, total)
	}
}
