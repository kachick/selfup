package runner

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const defaultPrefix string = " selfup "

func TestDryRun(t *testing.T) {
	type testCase struct {
		input  string
		prefix string
		skipBy string
		ok     bool
		want   Result
	}
	testCases := map[string]testCase{
		"Happy Path": {
			input: `Header
will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.39.0"] }
No JSON in this line
not_be_replacedB: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["echo", ":)"] }
`,
			prefix: defaultPrefix,
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
		"Another prefix": {
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
		}, "SkipBy": {
			input: `Header
will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
not_be_replacedA: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "0.76.9"] }
`,
			prefix: defaultPrefix,
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
		}, "Handle fields": {
			input: `will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "    supertool  0.76.9  "], "nth": 2 }
`,
			prefix: defaultPrefix,
			skipBy: "",
			ok:     true,
			want: Result{
				NewLines: []string{
					`will_be_replaced: '0.76.9' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "    supertool  0.76.9  "], "nth": 2 }`,
				},
				Targets: []Target{
					{LineNumber: 1, Extracted: "0.39.0", Replacer: "0.76.9", IsChanged: true},
				},
				ChangedCount: 1,
				Total:        1,
			},
		}, "Special delimiter": {
			input: `will_be_replaced: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "supertool:0.76.9"], "nth": 2, "delimiter": ":" }
`,
			prefix: defaultPrefix,
			skipBy: "",
			ok:     true,
			want: Result{
				NewLines: []string{
					`will_be_replaced: '0.76.9' # selfup { "extract": "\\d[^']+", "replacer": ["echo", "supertool:0.76.9"], "nth": 2, "delimiter": ":" }`,
				},
				Targets: []Target{
					{LineNumber: 1, Extracted: "0.39.0", Replacer: "0.76.9", IsChanged: true},
				},
				ChangedCount: 1,
				Total:        1,
			},
		}, "Command returns unupdatable string": {
			input: `broken_command: '0.39.0' # selfup { "extract": "\\d[^']+", "replacer": ["echo", ":)"] }
`,
			prefix: defaultPrefix,
			skipBy: "",
			ok:     false,
		}, "Command is not found": {
			input: `Header
broken: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["this_command_does_not_exist_so_raise_errors_and_do_not_update_this_file"] }
`,
			prefix: defaultPrefix,
			skipBy: "",
			ok:     false,
		}, "Broken JSON": {
			input: `Header
broken: ':<' # selfup {{ """" }
`,
			prefix: defaultPrefix,
			skipBy: "",
			ok:     false,
		}, "Prefer SkipBy rather than no command error": {
			input: `Header
broken: ':<' # selfup { "extract": ":[<\\)]", "replacer": ["this_command_does_not_exist_so_raise_errors_and_do_not_update_this_file"] }
`,
			prefix: defaultPrefix,
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
		}, "Prefer SkipBy rather than broken JSON error": {
			input: `Header
broken: ':<' # selfup {{ """" }
`,
			prefix: defaultPrefix,
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

	for what, tc := range testCases {
		t.Run(what, func(t *testing.T) {
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
