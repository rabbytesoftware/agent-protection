package scan_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestScanSkipsProseAndLocks(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		want []string
	}{
		{name: "markdown prose is skipped", path: "README.md", want: nil},
		{name: "plain text is skipped", path: "docs/notes.txt", want: nil},
		{name: "restructured text is skipped", path: "docs/index.rst", want: nil},
		{name: "a license is skipped", path: "LICENSE", want: nil},
		{name: "a suffixed license is skipped", path: "LICENSE-MIT", want: nil},
		{name: "a notice is skipped", path: "NOTICE", want: nil},
		{name: "an authors list is skipped", path: "AUTHORS", want: nil},
		{name: "go.sum is skipped", path: "go.sum", want: nil},
		{name: "a lock file is skipped", path: "yarn.lock", want: nil},
		{name: "a json lock file is skipped", path: "package-lock.json", want: nil},
		{name: "a worktree git file is skipped", path: ".git", want: nil},
		{name: "an unknown config is scanned", path: ".tool-versions", want: []string{".tool-versions:1"}},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			scanned(
				t,
				map[string]string{tc.path: "x\n"},
				lexeme.Language{Spans: wholeText},
			),
			tc.want,
		)
	}
}
