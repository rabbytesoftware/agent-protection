package scan_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/scan"
)

func TestScanSkipsIgnoredFiles(
	t *testing.T,
) {
	root := t.TempDir()
	files := map[string]string{
		".gitignore":  "secret.note\n*.weird\n",
		"a.note":      "x\n",
		"secret.note": "x\n",
		"b.weird":     "x\n",
	}
	for path, code := range files {
		write(
			t,
			filepath.Join(
				root,
				path,
			),
			code,
		)
	}

	err := exec.Command(
		"git",
		"-C",
		root,
		"init",
		"--quiet",
	).Run()
	if err != nil {
		t.Skipf(
			"git unavailable: %v",
			err,
		)
	}

	findings, err := scan.Scan(
		root,
		nil,
		noteLanguages(),
		lexeme.Language{Spans: wholeText},
	)
	if err != nil {
		t.Fatalf(
			"Scan() error = %v",
			err,
		)
	}

	assertLocations(
		t,
		"git-ignored files are not scanned",
		findings,
		[]string{".gitignore:1", "a.note:1"},
	)
}
