package runner

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDryRun(t *testing.T) {
	testCases := []struct {
		description string
		input       string
		prefix      string
		skipBy      string
		ok          bool
		want        Result
	}{
		{
			description: "Happy Path",
			input: `Header
will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.39.0"] }
No JSON in this line
not_be_replacedB: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["echo", ":)"] }
`,
			prefix: " selfup ",
			skipBy: "",
			ok:     true,
			want: Result{
				NewLines: []string{
					`Header`,
					`will_be_replaced: '0.76.9' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }`,
					`not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.39.0"] }`,
					`No JSON in this line`,
					`not_be_replacedB: ':)' # selfup { "extract": ":[<\\)]", "replacer": ["echo", ":)"] }`,
				},
				Targets: []Target{
					{LineNumber: 2, Extracted: "0.39.0", Replacer: "0.76.9", IsChanged: true},
					{LineNumber: 3, Extracted: "0.39.0", Replacer: "0.39.0"},
					{LineNumber: 5, Extracted: ":<", Replacer: ":)", IsChanged: true},
				},
				ChangedCount: 2,
				Total:        3,
			},
		},
		{
			description: "Another prefix",
			input: `Header
will_be_replaced: '0.39.0' // Update this line with { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
`,
			prefix: "// Update this line with ",
			skipBy: "",
			ok:     true,
			want: Result{
				NewLines: []string{
					`Header`,
					`will_be_replaced: '0.76.9' // Update this line with { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }`,
					`not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }`,
				},
				Targets: []Target{
					{LineNumber: 2, Extracted: "0.39.0", Replacer: "0.76.9", IsChanged: true},
				},
				ChangedCount: 1,
				Total:        1,
			},
		}, {
			description: "SkipBy",
			input: `Header
will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
`,
			prefix: " selfup ",
			skipBy: "not_be_replaced",
			ok:     true,
			want: Result{
				NewLines: []string{
					`Header`,
					`will_be_replaced: '0.76.9' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }`,
					`not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }`,
				},
				Targets: []Target{
					{LineNumber: 2, Extracted: "0.39.0", Replacer: "0.76.9", IsChanged: true},
				},
				ChangedCount: 1,
				Total:        1,
			},
		}, {
			description: "Command is not found",
			input: `Header
broken: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["this_command_does_not_exist_so_raise_errors_and_do_not_update_this_file"] }
`,
			prefix: " selfup ",
			skipBy: "",
			ok:     false,
		}, {
			description: "Broken JSON",
			input: `Header
broken: ':<' # selfup {{ """" }
`,
			prefix: " selfup ",
			skipBy: "",
			ok:     false,
		}, {
			description: "Prefer SkipBy rather than no command error",
			input: `Header
broken: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["this_command_does_not_exist_so_raise_errors_and_do_not_update_this_file"] }
`,
			prefix: " selfup ",
			skipBy: "this_command_does_not_exist",
			ok:     true,
			want: Result{
				NewLines: []string{
					`Header`,
					`broken: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["this_command_does_not_exist_so_raise_errors_and_do_not_update_this_file"] }`,
				},
				Targets:      []Target{},
				ChangedCount: 0,
				Total:        0,
			},
		}, {
			description: "Prefer SkipBy rather than broken JSON error",
			input: `Header
broken: ':<' # selfup {{ """" }
`,
			prefix: " selfup ",
			skipBy: "broken",
			ok:     true,
			want: Result{
				NewLines: []string{
					`Header`,
					`broken: ':<' # selfup {{ """" }`,
				},
				Targets:      []Target{},
				ChangedCount: 0,
				Total:        0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := DryRun(strings.NewReader(tc.input), tc.prefix, tc.skipBy)
			if err != nil {
				if tc.ok {
					t.Fatalf("unexpected error happened: %v", err)
				} else {
					return
				}
			}
			if !tc.ok {
				t.Fatalf("expected error did not happen")
			}

			if diff := cmp.Diff(tc.want, result); diff != "" {
				t.Errorf("wrong result: %s", diff)
			}
		})
	}
}
