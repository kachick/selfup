package updater

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/huandu/xstrings"
	"golang.org/x/xerrors"
)

type Definition struct {
	Regex  string
	Script string
}

// Returns new body and true if it is changed
func Update(path string, prefix string, isListMode bool, skipBy string) (string, bool, error) {
	newLines := []string{}
	isChanged := false

	file, err := os.Open(path)
	if err != nil {
		return "", false, xerrors.Errorf("%s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

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

		err = json.Unmarshal([]byte(tail), definition)
		if err != nil {
			return "", false, xerrors.Errorf("%s:%d: Unmarsharing `%s` as JSON has been failed, check the given prefix: %w", path, lineNumber, tail, err)
		}
		re := regexp.MustCompile(definition.Regex)
		if isListMode {
			fmt.Printf("%s:%d: %s\n", path, lineNumber, re.FindString(head))
			continue
		}
		out, err := exec.Command("bash", "-c", definition.Script).Output()
		if err != nil {
			return "", false, xerrors.Errorf("%s:%d: Executing %s with bash has been failed: %w", path, lineNumber, definition.Script, err)
		}
		replaced := re.ReplaceAllString(head, strings.TrimSuffix(string(out), "\n"))
		if !isChanged {
			isChanged = replaced != head
		}
		newLines = append(newLines, replaced+match+tail)
	}

	return strings.Join(newLines, "\n"), isChanged, nil
}
