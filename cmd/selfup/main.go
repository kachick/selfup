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
	prefixFlag := flag.String("prefix", "", "prefix to write json")
	listFlag := flag.Bool("list-targets", false, "print target lines without actual replacing")
	versionFlag := flag.Bool("version", false, "print the version of this program")

	const usage = `Usage: selfup [OPTIONS] [PATH]...

$ selfup --prefix='# selfup ' .github/workflows/*.yml
$ selfup --prefix='# selfup ' --list-targets .github/workflows/*.yml
`

	flag.Usage = func() {
		// https://github.com/golang/go/issues/57059#issuecomment-1336036866
		fmt.Printf("%s", usage+"\n\n")
		fmt.Println("Usage of command:")
		flag.PrintDefaults()
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

	prefix := *prefixFlag
	isListMode := *listFlag

	if prefix == "" {
		flag.Usage()
		log.Fatalf("%+v", xerrors.New("No prefix is specified"))
	}

	wg := &sync.WaitGroup{}
	for _, path := range flag.Args() {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			newBody, isDirty, err := updater.Update(path, prefix, isListMode)

			if err != nil {
				log.Fatalf("%+v", err)
			}

			if isDirty {
				err := os.WriteFile(path, []byte(newBody), os.ModePerm)
				if err != nil {
					log.Fatalf("%+v", xerrors.Errorf("%s: %w", path, err))
				}
			}
		}(path)
	}
	wg.Wait()
}
