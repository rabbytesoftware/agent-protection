package source_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestWalkerSkipsVendoredTreesOnlyAtTheRoot(
	t *testing.T,
) {
	root := t.TempDir()
	for _, path := range []string{"a/a.go", "vendor/v/v.go", ".git/g.go", "node_modules/n/n.go", "a/vendor/v.go", "a/node_modules/n.go"} {
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
	), filepath.Join(
		root,
		"a/node_modules",
	), filepath.Join(
		root,
		"a/vendor",
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

func TestHasSource(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"src/a.go",
		),
		"package a\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"tests/a_test.go",
		),
		"package a\n",
	)
	cases := map[string]bool{"src": true, "tests": false}
	for dir, want := range cases {
		hasSource, err := source.HasSource(filepath.Join(
			root,
			dir,
		))
		assertBoolResult(
			t,
			"HasSource("+dir+")",
			hasSource,
			err,
			want,
		)
	}

	_, err := source.HasSource(filepath.Join(
		root,
		"missing",
	))
	if err == nil {
		t.Error("HasSource(missing) error = nil, want a read error")
	}
}

func TestEntries(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"a.go",
		),
		"package a\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"a_test.go",
		),
		"package a\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"child/b.go",
		),
		"package b\n",
	)
	paths, err := source.Entries(root)
	if err != nil {
		t.Fatalf(
			"Entries() error = %v",
			err,
		)
	}

	if !slices.Equal(
		paths,
		[]string{filepath.Join(
			root,
			"a.go",
		)},
	) {
		t.Errorf(
			"Entries() = %v, want only a.go",
			paths,
		)
	}
}
