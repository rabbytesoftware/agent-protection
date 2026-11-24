package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

func TestRunnerRejectsFolderWithoutLanguageMarker(
	t *testing.T,
) {
	_, err := runner.Runner(t.TempDir())
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
		[]byte("package clean\n\n// Version is the fixture's release.\nconst Version = \"1\"\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write version.go: %v",
			err,
		)
	}
	output, err := runner.Runner(root)
	if err != nil {
		t.Fatalf(
			"Runner() error = %v",
			err,
		)
	}
	if len(output.Findings) != 0 {
		t.Errorf(
			"Runner() findings = %v, want none for a compliant module",
			output.Findings,
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
	output, err := runner.Runner(root)
	if err != nil {
		t.Fatalf(
			"Runner() with %s error = %v",
			marker,
			err,
		)
	}
	if len(output.Findings) != 0 {
		t.Errorf(
			"Runner() with %s findings = %v, want none",
			marker,
			output.Findings,
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
