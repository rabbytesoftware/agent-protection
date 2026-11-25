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
		[]string{
			"pkg/a.go",
			"pkg/doc.go",
			"pkg/gen.go",
			"pkg/late.go",
			"pkg/text.go",
		},
	) {
		t.Errorf(
			"Source() = %v, want every non-test source file",
			got,
		)
	}

	handWritten := relPaths(source.HandWritten(files))
	if !slices.Equal(
		handWritten,
		[]string{
			"pkg/a.go",
			"pkg/doc.go",
			"pkg/late.go",
			"pkg/text.go",
		},
	) {
		t.Errorf(
			"HandWritten() = %v, want every file but the generated one",
			handWritten,
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
		[]string{
			"pkg/a_test.go",
			"pkg/either_test.go",
			"pkg/gen_test.go",
			"pkg/unit_test.go",
		},
	) {
		t.Errorf(
			"Tests() = %v, want only the unit test files",
			got,
		)
	}
}

func TestTestsReportsMalformedBuildConstraints(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"broken_test.go",
		),
		"//go:build (integration\n\npackage p\n",
	)
	_, err := source.Tests(root)
	if err == nil {
		t.Error("Tests() error = nil, want a build constraint error")
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

	want := []string{}
	for _, name := range []string{
		"a.go",
		"doc.go",
		"gen.go",
		"late.go",
		"text.go",
	} {
		want = append(
			want,
			filepath.Join(
				root,
				"pkg",
				name,
			),
		)
	}

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
	cases := map[string]bool{
		"pkg/gen.go":  true,
		"pkg/a.go":    false,
		"pkg/doc.go":  false,
		"pkg/text.go": false,
		"pkg/late.go": false,
	}
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
		"// Code generated by sqlc. DO NOT EDIT.\n\npackage pkg\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"sqlc.yaml",
		),
		"version: \"2\"\nsql:\n  - gen:\n      go:\n        out: pkg\n",
	)
	for path, content := range handWrittenFixtures() {
		writeFile(
			t,
			filepath.Join(
				root,
				path,
			),
			content,
		)
	}

	return root
}

func handWrittenFixtures() (contents map[string]string) {
	return map[string]string{
		"pkg/it_test.go":       "//go:build integration\n\npackage pkg\n",
		"pkg/linux_it_test.go": "//go:build integration && linux\n\npackage pkg\n",
		"pkg/unit_test.go":     "//go:build !integration\n\npackage pkg\n",
		"pkg/either_test.go":   "//go:build integration || unit\n\npackage pkg\n",
		"pkg/doc.go":           "// Package pkg is hand written. DO NOT EDIT.\npackage pkg\n",
		"pkg/text.go":          "package pkg\n\nconst header = \"// Code generated by tool. DO NOT EDIT.\"\n",
		"pkg/late.go":          "package pkg\n\n// Code generated by tool. DO NOT EDIT.\n",
		"pkg/gen_test.go":      "// Code generated by tool. DO NOT EDIT.\n\npackage pkg\n",
	}
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

func writeFiles(
	t *testing.T,
	root string,
	files map[string]string,
) {
	t.Helper()
	for path, content := range files {
		writeFile(
			t,
			filepath.Join(
				root,
				path,
			),
			content,
		)
	}
}
