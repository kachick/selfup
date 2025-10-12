package runner

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
	NewLines     []string
	Targets      []Target
	ChangedCount int
	Total        int
}

// Like a ruby's String#partition
func partition(s string, sep *regexp.Regexp) (before string, separator string, after string, found bool) {
	location := sep.FindStringIndex(s)

	if location == nil {
		return "", "", "", false
	}

	before = s[:location[0]]
	separator = s[location[0]:location[1]]
	after = s[location[1]:]

	return before, separator, after, true
}

func DryRun(r io.Reader, prefix *regexp.Regexp, skipBy string) (Result, error) {
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
		headWithVersion, separator, jsonStr, found := partition(line, prefix)
		if !found || len(jsonStr) == 0 || jsonStr[0] != '{' {
			newLines = append(newLines, line)
			continue
		}

		def := &Definition{}
		totalCount += 1

		err := json.Unmarshal([]byte(jsonStr), def)
		if err != nil {
			return Result{}, xerrors.Errorf("%d: Unmarsharing `%s` as JSON has been failed, check the given prefix: %w", lineNumber, jsonStr, err)
		}
		extractor := regexp.MustCompile(def.Extract)
		if len(def.Command) < 1 {
			return Result{}, xerrors.Errorf("%d: Given JSON `%s` does not include commands", lineNumber, jsonStr)
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
			if def.Nth > len(fields) {
				return Result{}, xerrors.Errorf("%d: Accessing invalid fields: STDOUT:%s Delimiter:%s Nth:%d", lineNumber, cmdResult, def.Delimiter, def.Nth)
			}
			index := def.Nth - 1
			replacer = fields[index]
		}
		extracted := extractor.FindString(headWithVersion)
		replaced := strings.Replace(headWithVersion, extracted, replacer, 1)
		isChanged := false
		extractedToEnsure := extractor.FindString(replaced)
		if replacer != extractedToEnsure {
			return Result{}, xerrors.Errorf("%d: The result of updater command has malformed format: %s", lineNumber, replacer)
		}
		if replaced != headWithVersion {
			isChanged = true
			changedCount++
		}
		newLines = append(newLines, replaced+separator+jsonStr)
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
		NewLines:     newLines,
		Targets:      targets,
		Total:        totalCount,
		ChangedCount: changedCount,
	}, nil
}
