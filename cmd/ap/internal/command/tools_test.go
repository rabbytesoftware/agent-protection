package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/guide"
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/rulebook"
	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

func TestCheck(
	t *testing.T,
) {
	t.Chdir(t.TempDir())
	clean := fixtureModule(
		t,
		"clean",
		"package clean\n\nconst Version = \"1\"\n",
	)
	failing := fixtureModule(
		t,
		"failing",
		"package failing\n\n// Version is a comment.\nconst Version = \"1\"\n",
	)
	cases := []struct {
		root     string
		wantCode int
	}{
		{root: t.TempDir(), wantCode: findingsExitCode},
		{root: clean, wantCode: cleanExitCode},
		{root: failing, wantCode: findingsExitCode},
	}
	for _, tc := range cases {
		assertCode(
			t,
			"Check("+tc.root+")",
			realTools().Check(tc.root),
			tc.wantCode,
		)
	}
}

func TestCheckReportsUnwritableReport(
	t *testing.T,
) {
	directory := t.TempDir()
	t.Chdir(directory)
	err := os.Mkdir(
		filepath.Join(
			directory,
			reportPath,
		),
		0o755,
	)
	if err != nil {
		t.Fatal(err)
	}

	clean := fixtureModule(
		t,
		"clean",
		"package clean\n\nconst Version = \"1\"\n",
	)
	assertCode(
		t,
		"Check() with an unwritable report",
		realTools().Check(clean),
		findingsExitCode,
	)
}

func TestCheckReportsRegistryFailure(
	t *testing.T,
) {
	t.Chdir(t.TempDir())
	tools := realTools()
	tools.Docs = failingDocs
	assertCode(
		t,
		"Check() with a broken registry",
		tools.Check(fixtureModule(
			t,
			"clean",
			"package clean\n\nconst Version = \"1\"\n",
		)),
		findingsExitCode,
	)
}

func TestWriteReport(
	t *testing.T,
) {
	t.Chdir(t.TempDir())
	err := writeReport(report.Output{})
	if err != nil {
		t.Fatalf(
			"writeReport() error = %v",
			err,
		)
	}

	_, err = os.Stat(reportPath)
	if err != nil {
		t.Errorf(
			"report file missing: %v",
			err,
		)
	}
}

func TestExitCode(
	t *testing.T,
) {
	clean := exitCode(report.Output{})
	failing := exitCode(report.Output{Findings: []report.Finding{{Rule: "R03"}}})
	isCorrect := clean == cleanExitCode && failing == findingsExitCode
	if !isCorrect {
		t.Errorf(
			"exitCode() = %d clean and %d failing, want %d and %d",
			clean,
			failing,
			cleanExitCode,
			findingsExitCode,
		)
	}
}

func realTools() (tools Tools) {
	return Tools{
		Run:     runner.Runner,
		Rewrite: runner.Fix,
		Docs:    rulebook.Docs,
		Manual:  guide.Guide,
	}
}

func failingDocs() (docs map[string]report.RuleDoc, err error) {
	return nil,
		os.ErrNotExist
}

func assertCode(
	t *testing.T,
	call string,
	code int,
	wantCode int,
) {
	t.Helper()
	if code != wantCode {
		t.Errorf(
			"%s = %d, want %d",
			call,
			code,
			wantCode,
		)
	}
}

func fixtureModule(
	t *testing.T,
	name string,
	content string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"module example.com/"+name+"\n\ngo 1.24\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"version.go",
		),
		content,
	)
	return root
}

func writeFile(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.WriteFile(
		path,
		[]byte(content),
		reportFileMode,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFix(
	t *testing.T,
) {
	failing := fixtureModule(
		t,
		"failing",
		"package failing\n\nfunc F(a int, b int) {}\n",
	)
	cases := []struct {
		root     string
		wantCode int
	}{
		{root: failing, wantCode: cleanExitCode},
		{root: t.TempDir(), wantCode: findingsExitCode},
	}
	for _, tc := range cases {
		assertCode(
			t,
			"Fix("+tc.root+")",
			realTools().Fix(tc.root),
			tc.wantCode,
		)
	}
}

func TestGuide(
	t *testing.T,
) {
	assertCode(
		t,
		"Guide()",
		realTools().Guide(),
		cleanExitCode,
	)
}

func TestMarkdown(
	t *testing.T,
) {
	directory := t.TempDir()
	valid := filepath.Join(
		directory,
		"valid.json",
	)
	invalid := filepath.Join(
		directory,
		"invalid.json",
	)
	writeFile(
		t,
		valid,
		`{"rule_docs":{},"findings":[]}`,
	)
	writeFile(
		t,
		invalid,
		"not json",
	)
	cases := []struct {
		path     string
		wantCode int
	}{
		{path: valid, wantCode: cleanExitCode},
		{path: invalid, wantCode: findingsExitCode},
		{
			path: filepath.Join(
				directory,
				"missing.json",
			),
			wantCode: findingsExitCode,
		},
	}
	for _, tc := range cases {
		assertCode(
			t,
			"Markdown("+tc.path+")",
			realTools().Markdown(tc.path),
			tc.wantCode,
		)
	}
}
