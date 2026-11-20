package goast

import "testing"

func TestOnlyTypecheckIssues(t *testing.T) {
	testCases := []struct {
		name string
		out  string
		want bool
	}{
		{
			name: "only a typecheck error means the whole run aborted",
			out:  "pkg/x.go:1:1: something went wrong (typecheck)\n",
			want: true,
		},
		{
			name: "a typecheck error alongside real findings means other linters did run",
			out:  "pkg/x.go:1:1: something went wrong (typecheck)\npkg/y.go:2:2: cyclomatic complexity 9 (gocyclo)\n",
			want: false,
		},
		{
			name: "no typecheck error at all",
			out:  "pkg/y.go:2:2: cyclomatic complexity 9 (gocyclo)\n",
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := onlyTypecheckIssues(tc.out)
			if got != tc.want {
				t.Errorf("onlyTypecheckIssues() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  int
	}{
		{name: "two lines with trailing newline", input: "a\nb\n", want: 2},
		{name: "two lines without trailing newline", input: "a\nb", want: 2},
		{name: "empty string", input: "", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitLines(tc.input)
			if len(got) != tc.want {
				t.Errorf("splitLines(%q) = %d lines, want %d: %v", tc.input, len(got), tc.want, got)
			}
		})
	}
}
