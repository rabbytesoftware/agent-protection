package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lint"
)

func TestLintSkipsTagsItCannotEnumerate(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		taggedScript,
	)
	cases := []struct {
		name       string
		constraint string
		want       int
	}{
		{name: "a constraint over too many tags is left untouched", constraint: "a && b && c && d && e && f && g && h && i && j && k && l && m", want: 1},
		{name: "an unsatisfiable constraint needs no run", constraint: "prod && !prod", want: 1},
		{name: "a constraint satisfied only by its smallest tag set adds that set", constraint: "(a && b) || c", want: 2},
	}
	for _, tc := range cases {
		findings, err := lintSnippets(moduleWith(
			t,
			map[string]string{"x.go": "//go:build " + tc.constraint + "\n\npackage p\n"},
		))
		assertCount(
			t,
			tc.name,
			findings,
			err,
			tc.want,
		)
	}
}

func assertCount(
	t *testing.T,
	name string,
	findings []string,
	err error,
	want int,
) {
	t.Helper()
	isMismatch := err != nil ||
		len(findings) != want
	if isMismatch {
		t.Errorf(
			"%s: Lint() = %v, %v, want %d findings",
			name,
			findings,
			err,
			want,
		)
	}
}

func moduleWith(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	files["go.mod"] = "module example.com/p\n"
	for name, content := range files {
		writeFixture(
			t,
			filepath.Join(
				root,
				name,
			),
			content,
		)
	}

	return root
}

func lintSnippets(
	root string,
) (snippets []string, err error) {
	findings, err := lint.Lint(root)
	for _, finding := range findings {
		snippets = append(
			snippets,
			finding.Snippet,
		)
	}

	return snippets,
		err
}

func readFixture(
	t *testing.T,
	path string,
) (content string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}

	return string(data)
}
