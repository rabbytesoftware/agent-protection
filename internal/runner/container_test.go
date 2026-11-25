package runner_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

func TestRunnerRejectsFolderWithoutLanguageMarker(
	t *testing.T,
) {
	_, err := runner.Runner(
		t.TempDir(),
		os.Getenv,
	)
	if err == nil {
		t.Error("Runner() error = nil, want an unsupported-folder error")
	}
}

func TestRunnerChecksGoModule(
	t *testing.T,
) {
	root := t.TempDir()
	err := os.WriteFile(
		filepath.Join(
			root,
			"go.mod",
		),
		[]byte("module example.com/clean\n\ngo 1.24\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write go.mod: %v",
			err,
		)
	}

	err = os.WriteFile(
		filepath.Join(
			root,
			"version.go",
		),
		[]byte("package clean\n\nconst Version = \"1\"\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write version.go: %v",
			err,
		)
	}

	findings, err := runner.Runner(
		root,
		os.Getenv,
	)
	if err != nil {
		t.Fatalf(
			"Runner() error = %v",
			err,
		)
	}

	if len(findings) != 0 {
		t.Errorf(
			"Runner() findings = %v, want none for a compliant module",
			findings,
		)
	}
}

func TestRunnerScansCommentsInEveryLanguage(
	t *testing.T,
) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":     "module example.com/clean\n\n// pinned\ngo 1.24\n",
		"version.go": "package clean\n\n// Version is the release.\nconst Version = \"1\"\n",
		"release.sh": "#!/bin/sh\n# hack\ngo build ./...\n",
	}
	for path, content := range files {
		writeExecutable(
			t,
			filepath.Join(
				root,
				path,
			),
			content,
		)
	}

	findings, err := runner.Runner(
		root,
		os.Getenv,
	)
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	slices.Sort(got)
	want := []string{"R07 go.mod:3", "R07 release.sh:2", "R07 version.go:3"}
	isWrong := err != nil || !slices.Equal(
		got,
		want,
	)
	if isWrong {
		t.Errorf(
			"Runner() = %v, %v, want each comment reported once: %v",
			got,
			err,
			want,
		)
	}
}

func TestRunnerDispatchesByLanguageMarker(
	t *testing.T,
) {
	cases := []struct {
		marker   string
		content  string
		tool     string
		toolBody string
	}{
		{marker: "Cargo.toml", content: "[package]\nname = \"clean\"\n", tool: "cargo", toolBody: "exit 0"},
		{marker: "package.json", content: "{}\n", tool: "bun", toolBody: "printf '[]'"},
	}
	for _, tc := range cases {
		assertDispatch(
			t,
			tc.marker,
			tc.content,
			tc.tool,
			tc.toolBody,
		)
	}
}

func TestFix(
	t *testing.T,
) {
	module := t.TempDir()
	writeExecutable(
		t,
		filepath.Join(
			module,
			"go.mod",
		),
		"module example.com/fixture\n\ngo 1.24\n",
	)
	writeExecutable(
		t,
		filepath.Join(
			module,
			"pair.go",
		),
		"package fixture\n\nfunc pair(a int, b int) {\n}\n",
	)
	changed, err := runner.Fix(module)
	isMismatch := err != nil || len(changed) != 1
	if isMismatch {
		t.Errorf(
			"Fix() = %v, %v, want pair.go rewritten",
			changed,
			err,
		)
	}
}

func TestFixRejectsOtherLanguages(
	t *testing.T,
) {
	for _, marker := range []string{"Cargo.toml", "package.json"} {
		assertFixRefuses(
			t,
			marker,
		)
	}
}

func assertFixRefuses(
	t *testing.T,
	marker string,
) {
	t.Helper()
	root := t.TempDir()
	writeExecutable(
		t,
		filepath.Join(
			root,
			marker,
		),
		"",
	)
	_, err := runner.Fix(root)
	if err == nil {
		t.Errorf(
			"Fix() with %s error = nil, want ap fix to refuse it",
			marker,
		)
	}
}

func TestFixRejectsFolderWithoutLanguageMarker(
	t *testing.T,
) {
	_, err := runner.Fix(t.TempDir())
	if err == nil {
		t.Error("Fix() error = nil, want an unsupported-folder error")
	}
}

func assertDispatch(
	t *testing.T,
	marker string,
	content string,
	tool string,
	toolBody string,
) {
	t.Helper()
	root := t.TempDir()
	tools := t.TempDir()
	writeExecutable(
		t,
		filepath.Join(
			root,
			marker,
		),
		content,
	)
	writeExecutable(
		t,
		filepath.Join(
			tools,
			tool,
		),
		"#!/bin/sh\n"+toolBody+"\n",
	)
	t.Setenv(
		"HOME",
		t.TempDir(),
	)
	t.Setenv(
		"PATH",
		tools+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	findings, err := runner.Runner(
		root,
		os.Getenv,
	)
	if err != nil {
		t.Fatalf(
			"Runner() with %s error = %v",
			marker,
			err,
		)
	}

	if len(findings) != 0 {
		t.Errorf(
			"Runner() with %s findings = %v, want none",
			marker,
			findings,
		)
	}
}

func writeExecutable(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.WriteFile(
		path,
		[]byte(content),
		0o700,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}
