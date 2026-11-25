package workspace_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/workspace"
)

func TestWorkspace(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{name: "a single module is checked once", files: map[string]string{}, want: []string{"X a.go:1"}},
		{name: "a nested module without a go.work is reported", files: map[string]string{"internal/rounding/go.mod": "module example.com/rounding\n"}, want: []string{"X a.go:1", "R70 internal/rounding/go.mod"}},
		{name: "a nested module with an unreadable go.mod is still reported", files: map[string]string{"internal/rounding/go.mod": ""}, want: []string{"X a.go:1", "R70 internal/rounding/go.mod"}},
		{name: "a go.work member is checked as its own module with prefixed locations", files: map[string]string{"internal/rounding/go.mod": "module example.com/rounding\n", "go.work": "go 1.24\n\nuse (\n\t.\n\t./internal/rounding // the rounding module\n)\n"}, want: []string{"X a.go:1", "X internal/rounding/a.go:1", "X internal/rounding"}},
		{name: "a go.work listing other modules still reports the unlisted one", files: map[string]string{"a/go.mod": "module a\n", "b/go.mod": "module b\n", "go.work": "use ./a\n"}, want: []string{"X a.go:1", "X a/a.go:1", "X a", "R70 b/go.mod"}},
		{name: "vendored, testdata, hidden and underscore modules are ignored", files: map[string]string{"vendor/x/go.mod": "module x\n", "pkg/testdata/m/go.mod": "module m\n", ".cache/m/go.mod": "module m\n", "_old/go.mod": "module old\n", "node_modules/m/go.mod": "module m\n"}, want: []string{"X a.go:1"}},
	}
	for _, tc := range cases {
		root := moduleWith(
			t,
			tc.files,
		)
		assertWorkspace(
			t,
			tc.name,
			root,
			tc.want,
		)
	}
}

func TestWorkspaceSkipsGitIgnoredModules(
	t *testing.T,
) {
	root := moduleWith(
		t,
		map[string]string{
			".gitignore":           "build/\n",
			"build/out/go.mod":     "module out\n",
			"internal/kept/go.mod": "module kept\n",
		},
	)
	err := exec.Command(
		"git",
		"init",
		"-q",
		root,
	).Run()
	if err != nil {
		t.Skipf(
			"git unavailable: %v",
			err,
		)
	}

	assertWorkspace(
		t,
		"git-ignored modules are skipped",
		root,
		[]string{"X a.go:1", "R70 internal/kept/go.mod"},
	)
}

func TestWorkspaceErrors(
	t *testing.T,
) {
	failingMember := moduleWith(
		t,
		map[string]string{"m/go.mod": "module m\n", "go.work": "use ./m\n"},
	)
	cases := []struct {
		name  string
		root  string
		check func(string, func(string) string) ([]report.Finding, error)
	}{
		{name: "a failing root check is an error", root: t.TempDir(), check: failingBelow("")},
		{
			name: "a missing root is an error",
			root: filepath.Join(
				t.TempDir(),
				"missing",
			),
			check: located(""),
		},
		{name: "a failing member check is an error", root: failingMember, check: failingBelow(failingMember)},
	}
	for _, tc := range cases {
		_, err := workspace.Workspace(
			tc.root,
			os.Getenv,
			tc.check,
		)
		assertFailed(
			t,
			tc.name,
			err,
		)
	}
}

func assertWorkspace(
	t *testing.T,
	name string,
	root string,
	want []string,
) {
	t.Helper()
	findings, err := workspace.Workspace(
		root,
		os.Getenv,
		located(root),
	)
	got := summaries(findings)
	isMismatch := err != nil || !slices.Equal(
		got,
		want,
	)
	if isMismatch {
		t.Errorf(
			"%s: Workspace() = %v, %v, want %v",
			name,
			got,
			err,
			want,
		)
	}
}

func assertFailed(
	t *testing.T,
	name string,
	err error,
) {
	t.Helper()
	if err == nil {
		t.Errorf(
			"%s: Workspace() error = nil, want one",
			name,
		)
	}
}

func located(
	root string,
) (check func(string, func(string) string) ([]report.Finding, error)) {
	return func(
		directory string,
		environment func(string) string,
	) (findings []report.Finding, err error) {
		return moduleFindings(
				root,
				directory,
			),
			nil
	}
}

func moduleFindings(
	root string,
	directory string,
) (findings []report.Finding) {
	if directory == filepath.Clean(root) {
		return []report.Finding{{Rule: "X", Location: "a.go:1"}}
	}

	return []report.Finding{
		{Rule: "X", Location: "a.go:1"},
		{Rule: "X", Location: "."},
		{Rule: "X", Location: "a.go:1"},
	}
}

func failingBelow(
	root string,
) (check func(string, func(string) string) ([]report.Finding, error)) {
	return func(
		directory string,
		environment func(string) string,
	) (findings []report.Finding, err error) {
		return nil,
			failure(
				root,
				directory,
			)
	}
}

func failure(
	root string,
	directory string,
) (err error) {
	if directory == root {
		return nil
	}

	return errors.New("check failed")
}

func moduleWith(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	files["go.mod"] = "module example.com/root\n"
	for name, content := range files {
		writeFile(
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

func writeFile(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := errors.Join(
		os.MkdirAll(
			filepath.Dir(path),
			0o755,
		),
		os.WriteFile(
			path,
			[]byte(content),
			0o600,
		),
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}

func summaries(
	findings []report.Finding,
) (got []string) {
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	return got
}
