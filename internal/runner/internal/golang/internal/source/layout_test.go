package source_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestModulePath(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"// header\nmodule example.com/app\n\ngo 1.24\n",
	)
	modulePath, err := source.ModulePath(root)
	if err != nil {
		t.Fatalf(
			"ModulePath() error = %v",
			err,
		)
	}

	if modulePath != "example.com/app" {
		t.Errorf(
			"ModulePath() = %q, want example.com/app",
			modulePath,
		)
	}
}

func TestModulePathWithoutDirective(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"go 1.24\n",
	)
	_, err := source.ModulePath(root)
	if err == nil {
		t.Error("ModulePath() error = nil, want a missing-directive error")
	}
}

func TestIsDeclarationOnly(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"types.go",
		),
		"package p\n\ntype T interface{ M() }\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"behavior.go",
		),
		"package p\n\nfunc F() {}\n",
	)
	cases := map[string]bool{"types.go": true, "behavior.go": false}
	for name, want := range cases {
		declarationOnly, err := source.IsDeclarationOnly(filepath.Join(
			root,
			name,
		))
		assertBoolResult(
			t,
			"IsDeclarationOnly("+name+")",
			declarationOnly,
			err,
			want,
		)
	}

	_, err := source.IsDeclarationOnly(filepath.Join(
		root,
		"missing.go",
	))
	if err == nil {
		t.Error("IsDeclarationOnly(missing.go) error = nil, want a parse error")
	}
}

func TestIsInternalTree(
	t *testing.T,
) {
	cases := map[string]bool{"a/internal": true, "a/internal/b/c": true, "a/internalize": false, "a/b": false}
	for dir, want := range cases {
		assertBool(
			t,
			"IsInternalTree("+dir+")",
			source.IsInternalTree(dir),
			want,
		)
	}
}

func TestIsInternalDirectory(
	t *testing.T,
) {
	cases := map[string]bool{"a/internal": true, "a/internal/b": false}
	for dir, want := range cases {
		assertBool(
			t,
			"IsInternalDirectory("+dir+")",
			source.IsInternalDirectory(dir),
			want,
		)
	}
}

func TestIsSharedSupport(
	t *testing.T,
) {
	cases := map[string]bool{"internal/report": true, "internal/errors": true, "internal/runner": false}
	for dir, want := range cases {
		assertBool(
			t,
			"IsSharedSupport("+dir+")",
			source.IsSharedSupport(dir),
			want,
		)
	}
}

func TestIsMock(
	t *testing.T,
) {
	cases := map[string]bool{
		"internal/mocks/bus.go":  true,
		"store/fakes/clock.go":   true,
		"store/bus_mock.go":      true,
		"store/FakeClock.go":     true,
		"internal/bus/bus.go":    false,
		"internal/remote/sms.go": false,
	}
	for path, want := range cases {
		assertBool(
			t,
			"IsMock("+path+")",
			source.IsMock(path),
			want,
		)
	}
}

func TestIsCompositionRoot(
	t *testing.T,
) {
	cases := map[string]bool{"x/container.go": true, "x/wire.go": true, "x/arch.go": true, "x/helpers.go": false}
	for path, want := range cases {
		assertBool(
			t,
			"IsCompositionRoot("+path+")",
			source.IsCompositionRoot(
				path,
				"arch",
			),
			want,
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

func assertBool(
	t *testing.T,
	call string,
	got bool,
	want bool,
) {
	t.Helper()
	if got != want {
		t.Errorf(
			"%s = %v, want %v",
			call,
			got,
			want,
		)
	}
}

func assertBoolResult(
	t *testing.T,
	call string,
	got bool,
	err error,
	want bool,
) {
	t.Helper()
	if err != nil {
		t.Fatalf(
			"%s error = %v",
			call,
			err,
		)
	}

	assertBool(
		t,
		call,
		got,
		want,
	)
}
