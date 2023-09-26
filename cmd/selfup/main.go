package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	updater "github.com/kachick/selfup"
	"golang.org/x/xerrors"
)

var (
	// Used in goreleaser
	version = "dev"
	commit  = "none"

	revision = "rev"
)

func main() {
	sharedFlags := flag.NewFlagSet("run|list", flag.ExitOnError)
	prefixFlag := sharedFlags.String("prefix", "", "prefix to write json")
	skipByFlag := sharedFlags.String("skip-by", "", "skip to run if the line contains this string")
	versionFlag := flag.Bool("version", false, "print the version of this program")

	const usage = `Usage: selfup [SUB] [OPTIONS] [PATH]...

$ selfup run --prefix='# selfup ' .github/workflows/*.yml
$ selfup run --prefix='# selfup ' --skip-by='nix run' .github/workflows/*.yml
$ selfup list --prefix='# selfup ' .github/workflows/*.yml
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
	prefix := *prefixFlag
	skipBy := *skipByFlag

	if prefix == "" {
		flag.Usage()
		log.Fatalf("%+v", xerrors.New("No prefix is specified"))
	}

	wg := &sync.WaitGroup{}
	for _, path := range sharedFlags.Args() {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			newBody, isDirty, err := updater.Update(path, prefix, isListMode, skipBy)
			if err != nil {
				log.Fatalf("%+v", err)
			}

			if isRunMode && isDirty {
				err := os.WriteFile(path, []byte(newBody+"\n"), os.ModePerm)
				if err != nil {
					log.Fatalf("%+v", xerrors.Errorf("%s: %w", path, err))
				}
			}
		}(path)
	}
	wg.Wait()
}
