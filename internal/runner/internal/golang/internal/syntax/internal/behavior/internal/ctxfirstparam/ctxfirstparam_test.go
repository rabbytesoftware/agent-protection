package ctxfirstparam_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/ctxfirstparam"
)

func TestCtxfirstparam(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "context after the first parameter is reported at the parameter",
			code: "package p\n\nimport \"context\"\n\nfunc f(\n\ta int,\n\tctx context.Context,\n) {\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "context as the first parameter passes",
			code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n\ta int,\n) {\n}\n",
			want: nil,
		},
		{
			name: "context after the first parameter of an interface method is reported",
			code: "package p\n\nimport \"context\"\n\ntype runner interface {\n\tRun(\n\t\ta int,\n\t\tctx context.Context,\n\t)\n}\n",
			want: []string{"p.go:8"},
		},
		{
			name: "context stored in a struct field is reported",
			code: "package p\n\nimport \"context\"\n\ntype t struct {\n\tctx context.Context\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "function without parameters passes",
			code: "package p\n\nfunc f() {\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			ctxfirstparam.Ctxfirstparam([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		"p.go",
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
		Path:    "p.go",
		Rel:     "p.go",
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
