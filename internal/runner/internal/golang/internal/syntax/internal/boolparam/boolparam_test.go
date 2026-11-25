package boolparam_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/boolparam"
)

func TestBoolparam(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "function with a bool parameter is reported",
			code: "package p\n\nfunc f(\n\tcount int,\n\tvalidate bool,\n) {\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "method with a bool parameter is reported",
			code: "package p\n\ntype T struct{}\n\nfunc (t T) f(\n\tvalidate bool,\n) {\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "function without a bool parameter passes",
			code: "package p\n\nfunc f(\n\tcount int,\n\tlabel string,\n) {\n}\n",
			want: nil,
		},
		{
			name: "named type whose underlying type is bool is reported",
			code: "package p\n\ntype Express bool\n\nfunc f(\n\texpress Express,\n) {\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "named type derived from a named bool is reported",
			code: "package p\n\ntype Flag bool\n\ntype Express Flag\n\nfunc f(\n\texpress Express,\n) {\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "variadic bool parameter is reported",
			code: "package p\n\nfunc f(\n\tflags ...bool,\n) {\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "named type with a non-bool underlying type passes",
			code: "package p\n\ntype Mode int\n\nfunc f(\n\tmode Mode,\n) {\n}\n",
			want: nil,
		},
		{
			name: "struct type passes",
			code: "package p\n\ntype Options struct {\n\tfast bool\n}\n\nfunc f(\n\toptions Options,\n) {\n}\n",
			want: nil,
		},
		{
			name: "bool result passes",
			code: "package p\n\nfunc f(\n\tcount int,\n) (ok bool) {\n\treturn count > 0\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			boolparam.Boolparam([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestBoolparamAcrossFiles(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files []source.File
		want  []string
	}{
		{
			name: "named bool declared in another file of the package is reported",
			files: []source.File{
				parsedAt(
					t,
					"p/express.go",
					"package p\n\ntype Express bool\n",
				),
				parsedAt(
					t,
					"p/price.go",
					"package p\n\nfunc f(\n\texpress Express,\n) {\n}\n",
				),
			},
			want: []string{"p/price.go:3"},
		},
		{
			name: "named bool declared in another package passes",
			files: []source.File{
				parsedAt(
					t,
					"q/express.go",
					"package q\n\ntype Express bool\n",
				),
				parsedAt(
					t,
					"p/price.go",
					"package p\n\nfunc f(\n\texpress Express,\n) {\n}\n",
				),
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			boolparam.Boolparam(tc.files),
			tc.want,
		)
	}
}

func TestBoolparamSnippetNamesTheType(
	t *testing.T,
) {
	findings := boolparam.Boolparam([]source.File{parsed(
		t,
		"package p\n\ntype Express bool\n\nfunc f(\n\texpress Express,\n) {\n}\n",
	)})
	want := "func f(...) has a parameter of type Express, whose underlying type is bool: split it into two named functions or take a typed enum with named constants"
	isMismatch := len(findings) != 1 ||
		findings[0].Snippet != want
	if isMismatch {
		t.Errorf(
			"findings = %v, want one with snippet %q",
			findings,
			want,
		)
	}
}

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	return parsedAt(
		t,
		"p.go",
		code,
	)
}

func parsedAt(
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

	return source.File{
		Path:    path,
		Rel:     path,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}

func assertLocations(
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
			finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: locations = %v, want %v",
			name,
			got,
			want,
		)
	}
}
