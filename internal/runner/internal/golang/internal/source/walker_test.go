package source_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestWalkerSkipsVendoredTrees(
	t *testing.T,
) {
	root := t.TempDir()
	for _, path := range []string{"a/a.go", "vendor/v/v.go", ".git/g.go", "node_modules/n/n.go"} {
		writeFile(
			t,
			filepath.Join(
				root,
				path,
			),
			"package x\n",
		)
	}
	directories, err := source.Directories(root)
	if err != nil {
		t.Fatalf(
			"Directories() error = %v",
			err,
		)
	}
	want := []string{root, filepath.Join(
		root,
		"a",
	)}
	if !slices.Equal(
		directories,
		want,
	) {
		t.Errorf(
			"Directories() = %v, want %v",
			directories,
			want,
		)
	}
}

func writeFile(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"mkdir %s: %v",
			path,
			err,
		)
	}
	err = os.WriteFile(
		path,
		[]byte(content),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}
