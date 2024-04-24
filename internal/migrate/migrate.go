package migrate

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
)

type V1Schema struct {
	Extract   string   `json:"extract"`
	Command   []string `json:"replacer"`
	Nth       int      `json:"nth,omitempty"`
	Delimiter string   `json:"delimiter,omitempty"`
}

type BetaSchema struct {
	Regex  string `json:"regex"`
	Script string `json:"script"`
}

const defaultPrefix string = "# selfup "

func Migrate(path string) (bool, error) {
	newLines := []string{}
	isMigrated := false
	bytes, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(bytes)))
	for scanner.Scan() {
		line := scanner.Text()
		head, tail, found := strings.Cut(line, defaultPrefix)
		if !found {
			newLines = append(newLines, line)
			continue
		}

		beta := &BetaSchema{}
		err := json.Unmarshal([]byte(tail), beta)
		if err != nil {
			return false, err
		}
		if beta.Regex == "" || beta.Script == "" {
			newLines = append(newLines, line)
			continue
		}
		v1 := V1Schema{
			Extract:   beta.Regex,
			Command:   []string{"bash", "-c", beta.Script},
			Nth:       0,
			Delimiter: "",
		}
		migrated, err := json.MarshalIndent(v1, " ", "")
		if err != nil {
			return false, err
		}
		newLines = append(newLines, head+defaultPrefix+string(migrated))
		if !isMigrated {
			isMigrated = true
		}
	}
	if scanner.Err() != nil {
		return false, err
	}

	if isMigrated {
		err = os.WriteFile(path, []byte(strings.Join(newLines, "\n")), fs.ModePerm)
		if err != nil {
			return true, err
		}
	}

	return isMigrated, nil
}
