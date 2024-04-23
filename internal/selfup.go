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
	"github.com/huandu/xstrings"
	"golang.org/x/xerrors"
)

type Definition struct {
	Regex  string
	Script string
}

type Result struct {
	Lines   []string
	Changed int
	Total   int
}

// Returns new body and true if it is changed
func Update(path string, prefix string, isListMode bool, skipBy string, isColor bool) (Result, error) {
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
		head, match, tail := xstrings.LastPartition(line, prefix)
		if head == "" {
			newLines = append(newLines, line)
			continue
		}
		if skipBy != "" && strings.Contains(line, skipBy) {
			newLines = append(newLines, line)
			continue
		}

		definition := &Definition{}
		totalCount += 1

		err = json.Unmarshal([]byte(tail), definition)
		if err != nil {
			return Result{}, xerrors.Errorf("%s:%d: Unmarsharing `%s` as JSON has been failed, check the given prefix: %w", path, lineNumber, tail, err)
		}
		re := regexp.MustCompile(definition.Regex)
		out, err := exec.Command("bash", "-c", definition.Script).Output()
		if err != nil {
			return Result{}, xerrors.Errorf("%s:%d: Executing %s with bash has been failed: %w", path, lineNumber, definition.Script, err)
		}
		replacer := strings.TrimSuffix(string(out), "\n")
		extracted := re.FindString(head)
		estimation := " "
		suffix := ""
		if extracted != replacer {
			replacedCount += 1
			estimation = "✓"
			if isColor {
				replacer = green(replacer)
				estimation = green(estimation)
			}
			suffix = fmt.Sprintf(" => %s", replacer)
		}
		fmt.Println(fmt.Sprintf("%s %s:%d: %s", estimation, path, lineNumber, extracted) + suffix)

		if isListMode {
			continue
		}
		replaced := strings.Replace(head, extracted, replacer, 1)
		if !isChanged {
			isChanged = replaced != head
		}
		newLines = append(newLines, replaced+match+tail)
	}

	return Result{
		Lines:   newLines,
		Changed: replacedCount,
		Total:   totalCount,
	}, nil
}
