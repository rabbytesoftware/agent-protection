package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestRun(
	t *testing.T,
) {
	t.Chdir(t.TempDir())
	clean := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			clean,
			"go.mod",
		),
		"module example.com/clean\n\ngo 1.24\n",
	)
	writeFile(
		t,
		filepath.Join(
			clean,
			"version.go",
		),
		"package clean\n\nconst Version = \"1\"\n",
	)
	failing := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			failing,
			"go.mod",
		),
		"module example.com/failing\n\ngo 1.24\n",
	)
	writeFile(
		t,
		filepath.Join(
			failing,
			"version.go",
		),
		"package failing\n\n// Version is a comment.\nconst Version = \"1\"\n",
	)
	cases := []struct {
		arguments []string
		wantCode  int
	}{
		{arguments: []string{"ap"}, wantCode: usageExitCode},
		{arguments: []string{"ap", "guide"}, wantCode: cleanExitCode},
		{arguments: []string{"ap", "guide", "extra"}, wantCode: usageExitCode},
		{arguments: []string{"ap", "check", t.TempDir()}, wantCode: findingsExitCode},
		{arguments: []string{"ap", "check", clean}, wantCode: cleanExitCode},
		{arguments: []string{"ap", "check", failing}, wantCode: findingsExitCode},
		{arguments: []string{"ap", "markdown", reportPath}, wantCode: cleanExitCode},
		{arguments: []string{"ap", "fix", failing}, wantCode: cleanExitCode},
		{arguments: []string{"ap", "fix", t.TempDir()}, wantCode: findingsExitCode},
	}
	for _, tc := range cases {
		assertRun(
			t,
			tc.arguments,
			tc.wantCode,
		)
	}
}

func TestCommandArguments(
	t *testing.T,
) {
	cases := []struct {
		arguments   []string
		wantCommand string
		wantTarget  string
		wantErr     bool
	}{
		{arguments: []string{"ap", "check", "."}, wantCommand: "check", wantTarget: "."},
		{arguments: []string{"ap", "markdown", "ap-report.json"}, wantCommand: "markdown", wantTarget: "ap-report.json"},
		{arguments: []string{"ap", "fix", "."}, wantCommand: "fix", wantTarget: "."},
		{arguments: []string{"ap", "check"}, wantErr: true},
		{arguments: []string{"ap", "lint", "."}, wantErr: true},
		{arguments: []string{"ap", "check", ".", "extra"}, wantErr: true},
	}
	for _, tc := range cases {
		assertCommandArguments(
			t,
			tc.arguments,
			tc.wantCommand,
			tc.wantTarget,
			tc.wantErr,
		)
	}
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
	missing := filepath.Join(
		directory,
		"missing.json",
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
		{path: missing, wantCode: findingsExitCode},
	}
	for _, tc := range cases {
		assertMarkdownCode(
			t,
			tc.path,
			tc.wantCode,
		)
	}
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

func assertRun(
	t *testing.T,
	arguments []string,
	wantCode int,
) {
	t.Helper()
	code := run(arguments)
	if code != wantCode {
		t.Errorf(
			"run(%v) = %d, want %d",
			arguments,
			code,
			wantCode,
		)
	}
}

func assertCommandArguments(
	t *testing.T,
	arguments []string,
	wantCommand string,
	wantTarget string,
	wantErr bool,
) {
	t.Helper()
	command, target, err := commandArguments(arguments)
	isMismatch := command != wantCommand || target != wantTarget || (err != nil) != wantErr
	if isMismatch {
		t.Errorf(
			"commandArguments(%v) = %q, %q, %v; want %q, %q, error %v",
			arguments,
			command,
			target,
			err,
			wantCommand,
			wantTarget,
			wantErr,
		)
	}
}

func assertMarkdownCode(
	t *testing.T,
	path string,
	wantCode int,
) {
	t.Helper()
	code := markdown(path)
	if code != wantCode {
		t.Errorf(
			"markdown(%q) = %d, want %d",
			path,
			code,
			wantCode,
		)
	}
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
