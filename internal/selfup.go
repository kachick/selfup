package updater

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/xerrors"
)

type Definition struct {
	Extract   string   `json:"extract"`
	Command   []string `json:"replacer"`
	Nth       int      `json:"nth"`
	Delimiter string   `json:"delimiter"`
}

type Result struct {
	Lines   []string
	Changed int
	Total   int
}

// Returns new body and true if it is changed
func Update(path string, prefix string, skipBy string, isColor bool) (Result, error) {
	green := color.New(color.FgGreen).SprintFunc()
	newLines := []string{}
	isChanged := false

	file, err := os.Open(path)
	if err != nil {
		return Result{}, xerrors.Errorf("%s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	totalCount := 0
	replacedCount := 0

	for scanner.Scan() {
		lineNumber += 1
		line := scanner.Text()
		if skipBy != "" && strings.Contains(line, skipBy) {
			newLines = append(newLines, line)
			continue
		}
		head, tail, found := strings.Cut(line, prefix)
		if !found {
			newLines = append(newLines, line)
			continue
		}

		def := &Definition{}
		totalCount += 1

		err = json.Unmarshal([]byte(tail), def)
		if err != nil {
			return Result{}, xerrors.Errorf("%s:%d: Unmarsharing `%s` as JSON has been failed, check the given prefix: %w", path, lineNumber, tail, err)
		}
		re := regexp.MustCompile(def.Extract)
		if len(def.Command) < 1 {
			return Result{}, xerrors.Errorf("%s:%d: Given JSON `%s` does not include commands", path, lineNumber, tail)
		}
		cmd := def.Command[0]
		args := def.Command[1:]
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			return Result{}, xerrors.Errorf("%s:%d: Executing %s with bash has been failed: %w", path, lineNumber, cmd, err)
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
		if !isChanged {
			isChanged = replaced != head
		}
		extractedToEnsure := re.FindString(replaced)
		if replacer != extractedToEnsure {
			return Result{}, xerrors.Errorf("%s:%d: The result of updater command has malformed format: %s", path, lineNumber, replacer)
		}
		estimation := " "
		suffix := ""
		if replaced != head {
			replacedCount += 1
			estimation = "✓"
			if isColor {
				replacer = green(replacer)
				estimation = green(estimation)
			}
			suffix = fmt.Sprintf(" => %s", replacer)
		}
		newLines = append(newLines, replaced+prefix+tail)
		fmt.Println(fmt.Sprintf("%s %s:%d: %s", estimation, path, lineNumber, extracted) + suffix)
	}

	return Result{
		Lines:   newLines,
		Changed: replacedCount,
		Total:   totalCount,
	}, nil
}
