package noelse_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/noelse"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestNoelse(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "else branch is reported at the else keyword's block",
			code: "package p\n\nfunc f(\n\tx bool,\n) {\n\tif x {\n\t\treturn\n\t} else {\n\t\treturn\n\t}\n}\n",
			want: []string{"p.go:8"},
		},
		{
			name: "else if chain is reported",
			code: "package p\n\nfunc f(\n\tx int,\n) {\n\tif x == 1 {\n\t\treturn\n\t} else if x == 2 {\n\t\treturn\n\t}\n}\n",
			want: []string{"p.go:8"},
		},
		{
			name: "guard clause passes",
			code: "package p\n\nfunc f(\n\tx bool,\n) {\n\tif x {\n\t\treturn\n\t}\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			noelse.Noelse([]source.File{parsed(
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

	return source.File{Path: "p.go", Rel: "p.go", FileSet: fileSet, Syntax: syntax}
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
