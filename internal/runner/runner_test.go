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
