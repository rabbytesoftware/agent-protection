package source_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestSource(
	t *testing.T,
) {
	root := fixtureModule(t)
	files, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"Source() error = %v",
			err,
		)
	}
	got := relPaths(files)
	if !slices.Equal(
		got,
		[]string{"pkg/a.go"},
	) {
		t.Errorf(
			"Source() = %v, want only the hand-written source file",
			got,
		)
	}
}

func TestTests(
	t *testing.T,
) {
	root := fixtureModule(t)
	files, err := source.Tests(root)
	if err != nil {
		t.Fatalf(
			"Tests() error = %v",
			err,
		)
	}
	got := relPaths(files)
	if !slices.Equal(
		got,
		[]string{"pkg/a_test.go"},
	) {
		t.Errorf(
			"Tests() = %v, want only the unit test file",
			got,
		)
	}
}

func TestPaths(
	t *testing.T,
) {
	root := fixtureModule(t)
	paths, err := source.Paths(root)
	if err != nil {
		t.Fatalf(
			"Paths() error = %v",
			err,
		)
	}
	want := []string{filepath.Join(
		root,
		"pkg/a.go",
	), filepath.Join(
		root,
		"pkg/gen.go",
	)}
	if !slices.Equal(
		paths,
		want,
	) {
		t.Errorf(
			"Paths() = %v, want %v",
			paths,
			want,
		)
	}
}

func TestSourceReportsUnparsableFiles(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"broken.go",
		),
		"package\n",
	)
	_, err := source.Source(root)
	if err == nil {
		t.Error("Source() error = nil, want a parse error")
	}
}

func TestSourceReportsUnreadableFiles(
	t *testing.T,
) {
	root := t.TempDir()
	path := filepath.Join(
		root,
		"locked.go",
	)
	writeFile(
		t,
		path,
		"package locked\n",
	)
	err := os.Chmod(
		path,
		0,
	)
	if err != nil {
		t.Fatalf(
			"chmod: %v",
			err,
		)
	}

	_, err = source.Source(root)
	if err == nil {
		t.Error("Source() error = nil, want a read error")
	}
}

func TestParse(
	t *testing.T,
) {
	content := []byte("package p\n")
	file, err := source.Parse(
		"root",
		"root/dir/p.go",
		content,
	)
	if err != nil {
		t.Fatalf(
			"Parse() error = %v",
			err,
		)
	}

	isMismatch := file.Rel != "dir/p.go" || string(file.Content) != string(content)
	if isMismatch {
		t.Errorf(
			"Parse() = Rel %q Content %q, want dir/p.go and the parsed content",
			file.Rel,
			file.Content,
		)
	}

	_, err = source.Parse(
		"root",
		"root/dir/p.go",
		[]byte("package\n"),
	)
	if err == nil {
		t.Error("Parse() error = nil, want a parse error")
	}
}

func TestRelPath(
	t *testing.T,
) {
	rel, err := source.RelPath(
		"/repo",
		"/repo/internal/a.go",
	)
	isMismatch := err != nil || rel != "internal/a.go"
	if isMismatch {
		t.Errorf(
			"RelPath() = %q, %v, want internal/a.go, nil",
			rel,
			err,
		)
	}
}

func TestRelPathRejectsUnrelatablePaths(
	t *testing.T,
) {
	_, err := source.RelPath(
		"/repo",
		"relative/a.go",
	)
	if err == nil {
		t.Error("RelPath(/repo, relative/a.go) error = nil, want one")
	}
}

func TestWalkersRejectMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	walkers := []func(string) error{
		sourceError,
		testsError,
		pathsError,
		directoriesError,
	}
	for _, walker := range walkers {
		assertWalkError(
			t,
			walker,
			missing,
		)
	}
}

func assertWalkError(
	t *testing.T,
	walker func(string) error,
	root string,
) {
	t.Helper()
	err := walker(root)
	if err == nil {
		t.Errorf(
			"walking %s: error = nil, want one",
			root,
		)
	}
}

func sourceError(
	root string,
) (err error) {
	_, err = source.Source(root)
	return err
}

func testsError(
	root string,
) (err error) {
	_, err = source.Tests(root)
	return err
}

func pathsError(
	root string,
) (err error) {
	_, err = source.Paths(root)
	return err
}

func directoriesError(
	root string,
) (err error) {
	_, err = source.Directories(root)
	return err
}

func TestIsGenerated(
	t *testing.T,
) {
	root := fixtureModule(t)
	cases := map[string]bool{"pkg/gen.go": true, "pkg/a.go": false}
	for path, want := range cases {
		assertGenerated(
			t,
			filepath.Join(
				root,
				path,
			),
			want,
		)
	}
}

func TestIsGeneratedRejectsUnreadableFiles(
	t *testing.T,
) {
	_, err := source.IsGenerated(filepath.Join(
		t.TempDir(),
		"missing.go",
	))
	if err == nil {
		t.Error("IsGenerated(missing.go) error = nil, want a read error")
	}
}

func TestWalkersRejectUnreadableFiles(
	t *testing.T,
) {
	walkers := map[string]func(string) error{"broken.go": sourceError, "broken_test.go": testsError}
	for name, walker := range walkers {
		root := t.TempDir()
		symlink(
			t,
			filepath.Join(
				root,
				"missing.go",
			),
			filepath.Join(
				root,
				name,
			),
		)
		assertWalkError(
			t,
			walker,
			root,
		)
	}
}

func assertGenerated(
	t *testing.T,
	path string,
	want bool,
) {
	t.Helper()
	got, err := source.IsGenerated(path)
	isMismatch := err != nil || got != want
	if isMismatch {
		t.Errorf(
			"IsGenerated(%s) = %v, %v, want %v, nil",
			path,
			got,
			err,
			want,
		)
	}
}

func symlink(
	t *testing.T,
	target string,
	link string,
) {
	t.Helper()
	err := os.Symlink(
		target,
		link,
	)
	if err != nil {
		t.Fatalf(
			"symlink %s: %v",
			link,
			err,
		)
	}
}

func fixtureModule(
	t *testing.T,
) (root string) {
	t.Helper()
	root = t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"pkg/a.go",
		),
		"package pkg\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"pkg/a_test.go",
		),
		"package pkg\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"pkg/gen.go",
		),
		"// Code generated by tool. DO NOT EDIT.\n\npackage pkg\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"pkg/it_test.go",
		),
		"//go:build integration\n\npackage pkg\n",
	)
	return root
}

func relPaths(
	files []source.File,
) (paths []string) {
	for _, file := range files {
		paths = append(
			paths,
			file.Rel,
		)
	}
	return paths
}
