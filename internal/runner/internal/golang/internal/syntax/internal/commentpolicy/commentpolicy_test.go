package commentpolicy_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestCommentpolicy(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{
			name: "doc comment on an exported func is reported",
			path: "p.go",
			code: "package p\n\n// F does a thing.\nfunc F() {\n}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "inline comment is reported",
			path: "p.go",
			code: "package p\n\nfunc F() {\n\tF() // again\n}\n",
			want: []string{"R07 p.go:4"},
		},
		{
			name: "block comment is reported",
			path: "p.go",
			code: "package p\n\n/* header */\nfunc F() {\n}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "license header is reported",
			path: "p.go",
			code: "// Copyright the authors.\n\npackage p\n",
			want: []string{"R07 p.go:1"},
		},
		{
			name: "nolint suppression is reported",
			path: "p.go",
			code: "package p\n\nfunc F() {\n\tF() //nolint:errcheck\n}\n",
			want: []string{"R07 p.go:4"},
		},
		{
			name: "deprecation note is reported",
			path: "p.go",
			code: "package p\n\n// Deprecated: use G.\nfunc F() {\n}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "comment on a test function is reported",
			path: "p_test.go",
			code: "package p_test\n\nimport \"testing\"\n\n// TestF checks F.\nfunc TestF(\n\tt *testing.T,\n) {\n}\n",
			want: []string{"R07 p_test.go:5"},
		},
		{
			name: "compiler directives pass",
			path: "p.go",
			code: "//go:build linux\n\npackage p\n\nimport _ \"embed\"\n\n//go:embed p.txt\nvar text string\n\n//go:generate stringer -type=T\n//line p.go:1\n//export F\nfunc F() {\n}\n",
			want: nil,
		},
		{
			name: "prose mixed into a directive group is reported",
			path: "p.go",
			code: "package p\n\n// F is embedded.\n//go:noinline\nfunc F() {\n}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "cgo preamble passes",
			path: "p.go",
			code: "package p\n\n// #include <stdio.h>\nimport \"C\"\n",
			want: nil,
		},
		{
			name: "example output passes",
			path: "p_test.go",
			code: "package p_test\n\nimport \"fmt\"\n\nfunc ExampleF() {\n\tfmt.Println(1)\n\t// Output: 1\n}\n",
			want: nil,
		},
		{
			name: "output marker outside an example is reported",
			path: "p_test.go",
			code: "package p_test\n\nfunc helper() {\n\t// Output: 1\n}\n",
			want: []string{"R07 p_test.go:4"},
		},
		{
			name: "output marker in a source file is reported",
			path: "p.go",
			code: "package p\n\nfunc ExampleF() {\n\t// Output: 1\n}\n",
			want: []string{"R07 p.go:4"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			commentpolicy.Commentpolicy([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsed(
	t *testing.T,
	path string,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		path,
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return source.File{Path: path, Rel: path, FileSet: fileSet, Syntax: syntax}
}

func assertFindings(
	t *testing.T,
	name string,
	findings []report.Finding,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
