package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMigrate(t *testing.T) {
	type testCase struct {
		input         string
		expected      string
		shouldMigrate bool
		shouldFail    bool
	}

	testCases := map[string]testCase{
		"migrate beta schema to v1 schema": {
			input:         `# selfup {"regex": "foo", "script": "bar"}`,
			expected:      `# selfup {"extract":"foo","replacer":["bash","-c","bar"]}`,
			shouldMigrate: true,
			shouldFail:    false,
		},
		"no migration needed": {
			input:         `# selfup {"extract":"foo","replacer":["bash","-c","bar"],"nth":0,"delimiter":""}`,
			expected:      `# selfup {"extract":"foo","replacer":["bash","-c","bar"],"nth":0,"delimiter":""}`,
			shouldMigrate: false,
			shouldFail:    false,
		},
		"invalid json": {
			input:         `# selfup {"regex": "foo", "script": "bar"`,
			expected:      `# selfup {"regex": "foo", "script": "bar"`,
			shouldMigrate: false,
			shouldFail:    true,
		},
		"empty regex and script": {
			input:         `# selfup {"regex": "", "script": ""}`,
			expected:      `# selfup {"regex": "", "script": ""}`,
			shouldMigrate: false,
			shouldFail:    false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, strings.ReplaceAll(desc, " ", "-")+".txt")
			err := os.WriteFile(tmpFile, []byte(tc.input), 0644)
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}

			isMigrated, err := Migrate(tmpFile)

			if (err != nil) != tc.shouldFail {
				t.Fatalf("Unexpected error: %v, expectError: %v", err, tc.shouldFail)
			}

			if isMigrated != tc.shouldMigrate {
				t.Errorf("Unexpected isMigrated: %v, expected: %v", isMigrated, tc.shouldMigrate)
			}
			out, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read temporary file: %v", err)
			}
			actual := string(out)

			if diff := cmp.Diff(tc.expected, strings.TrimSpace(actual)); diff != "" {
				t.Errorf("wrong result: %s\nexpected: %s\nactual: %s", diff, tc.expected, actual)
			}
		})
	}
}
