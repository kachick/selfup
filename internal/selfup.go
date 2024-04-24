package updater

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

type Definition struct {
	Extract   string   `json:"extract"`
	Command   []string `json:"replacer"`
	Nth       int      `json:"nth"`
	Delimiter string   `json:"delimiter"`
}

type Target struct {
	LineNumber int
	Extracted  string
	Replacer   string
	IsChanged  bool
}

type Result struct {
	Lines        []string
	Targets      []Target
	ChangedCount int
	Total        int
}

func DryRun(r io.Reader, prefix string, skipBy string) (Result, error) {
	newLines := []string{}
	targets := []Target{}

	scanner := bufio.NewScanner(r)
	lineNumber := 0
	totalCount := 0
	changedCount := 0

	for scanner.Scan() {
		lineNumber += 1
		line := scanner.Text()
		if skipBy != "" && strings.Contains(line, skipBy) {
			newLines = append(newLines, line)
			continue
		}
		head, tail, found := strings.Cut(line, prefix+"{")
		if !found {
			newLines = append(newLines, line)
			continue
		}
		tail = "{" + tail

		def := &Definition{}
		totalCount += 1

		err := json.Unmarshal([]byte(tail), def)
		if err != nil {
			return Result{}, xerrors.Errorf("%d: Unmarsharing `%s` as JSON has been failed, check the given prefix: %w", lineNumber, tail, err)
		}
		re := regexp.MustCompile(def.Extract)
		if len(def.Command) < 1 {
			return Result{}, xerrors.Errorf("%d: Given JSON `%s` does not include commands", lineNumber, tail)
		}
		cmd := def.Command[0]
		args := def.Command[1:]
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			return Result{}, xerrors.Errorf("%d: Executing %s has been failed: %w", lineNumber, cmd, err)
		}
		cmdResult := strings.TrimSuffix(string(out), "\n")
		replacer := cmdResult
		if def.Nth > 0 {
			var fields []string
			if def.Delimiter == "" {
				fields = strings.Fields(cmdResult)
			} else {
				fields = strings.Split(cmdResult, def.Delimiter)
			}
			replacer = fields[def.Nth-1]
		}
		extracted := re.FindString(head)
		replaced := strings.Replace(head, extracted, replacer, 1)
		isChanged := false
		extractedToEnsure := re.FindString(replaced)
		if replacer != extractedToEnsure {
			return Result{}, xerrors.Errorf("%d: The result of updater command has malformed format: %s", lineNumber, replacer)
		}
		if replaced != head {
			isChanged = true
			changedCount++
		}
		newLines = append(newLines, replaced+prefix+tail)
		targets = append(targets, Target{
			LineNumber: lineNumber,
			Extracted:  extracted,
			Replacer:   replacer,
			IsChanged:  isChanged,
		})
	}

	err := scanner.Err()
	if err != nil {
		return Result{}, err
	}

	return Result{
		Lines:        newLines,
		Targets:      targets,
		Total:        totalCount,
		ChangedCount: changedCount,
	}, nil
}
