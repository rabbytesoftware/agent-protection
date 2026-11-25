package source_test

import (
	"fmt"
	"path/filepath"
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
	cases := map[string]bool{
		"a/internal":     true,
		"a/internal/b/c": true,
		"a/internalize":  false,
		"a/b":            false,
	}
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

func TestIsMock(
	t *testing.T,
) {
	cases := map[string]bool{
		"internal/mocks/bus.go":   true,
		"store/fakes/clock.go":    true,
		"store/bus_mock.go":       true,
		"store/fake_clock.go":     true,
		"store/mocks.go":          true,
		"store/FakeClock.go":      false,
		"internal/mockbuses/b.go": false,
		"internal/bus/bus.go":     false,
		"internal/remote/sms.go":  false,
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

func TestIsExample(
	t *testing.T,
) {
	cases := map[string]bool{"example/commands/c.go": true, "examples/x.go": true, "internal/exampler/x.go": false}
	for path, want := range cases {
		assertBool(
			t,
			"IsExample("+path+")",
			source.IsExample(path),
			want,
		)
	}
}

func TestIsModule(
	t *testing.T,
) {
	sources := []string{"a/internal/b", "c", "d/internalized"}
	cases := map[string]bool{
		"a":            true,
		"c":            false,
		"d":            false,
		"a/internal/b": false,
	}
	for dir, want := range cases {
		assertBool(
			t,
			"IsModule("+dir+")",
			source.IsModule(
				dir,
				sources,
			),
			want,
		)
	}
}

func TestIsCompositionRoot(
	t *testing.T,
) {
	cases := map[string]bool{
		"x/container.go": true,
		"x/main.go":      false,
		"x/wire.go":      false,
		"x/arch.go":      false,
		"x/helpers.go":   false,
	}
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

func TestBuildConstraint(
	t *testing.T,
) {
	cases := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{name: "a constrained file yields its expression", content: "//go:build prod && !windows\n\npackage p\n", want: "prod && !windows"},
		{name: "an unconstrained file yields nothing", content: "package p\n", want: "<nil>"},
		{name: "a malformed constraint is an error", content: "//go:build prod &&\n\npackage p\n", want: "<nil>", wantErr: true},
		{name: "an unparsable file is an error", content: "package\n", want: "<nil>", wantErr: true},
	}
	for _, tc := range cases {
		path := filepath.Join(
			t.TempDir(),
			"p.go",
		)
		writeFile(
			t,
			path,
			tc.content,
		)
		expression, err := source.BuildConstraint(path)
		assertConstraint(
			t,
			tc.name,
			fmt.Sprint(expression),
			err,
			tc.want,
			tc.wantErr,
		)
	}
}

func assertConstraint(
	t *testing.T,
	name string,
	got string,
	err error,
	want string,
	wantErr bool,
) {
	t.Helper()
	isMismatch := got != want ||
		(err != nil) != wantErr
	if isMismatch {
		t.Errorf(
			"%s: BuildConstraint() = %q, %v",
			name,
			got,
			err,
		)
	}
}
