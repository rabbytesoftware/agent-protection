package scan_test

import (
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/scan"
)

func TestWalker(
	t *testing.T,
) {
	cases := []struct {
		name     string
		path     string
		excluded []string
		want     []string
	}{
		{name: "nested source directory is scanned", path: "src/deep/a.note", want: []string{"src/deep/a.note:1"}},
		{name: "hidden directories are scanned", path: ".github/workflows/a.note", want: []string{".github/workflows/a.note:1"}},
		{name: "build output directories are scanned", path: "dist/a.note", want: []string{"dist/a.note:1"}},
		{name: "vendor is skipped", path: "vendor/a.note", want: nil},
		{name: "node_modules is skipped", path: "node_modules/p/a.note", want: nil},
		{name: "git directory is skipped", path: ".git/a.note", want: nil},
		{name: "testdata is skipped", path: "pkg/testdata/a.note", want: nil},
		{name: "a directory merely containing a skipped name is scanned", path: "vendored/a.note", want: []string{"vendored/a.note:1"}},
		{name: "an unmatched file is not scanned", path: "a.other", want: nil},
		{name: "an excluded extension is not scanned", path: "a.note", excluded: []string{noteExtension}, want: nil},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc.name,
			map[string]string{tc.path: "x\n"},
			tc.excluded,
			tc.want,
		)
	}
}

func TestWalkerFirstLanguageWins(
	t *testing.T,
) {
	root := t.TempDir()
	write(
		t,
		filepath.Join(
			root,
			"a.note",
		),
		"x\n",
	)
	findings, err := scan.Scan(
		root,
		nil,
		[]lexeme.Language{
			{Matches: isNote, Spans: noSpans},
			{Matches: isNote, Spans: wholeText},
		},
	)
	isWrong := err != nil || len(findings) != 0
	if isWrong {
		t.Errorf(
			"Scan() = %v, %v, want no findings from the first matching language",
			findings,
			err,
		)
	}
}
