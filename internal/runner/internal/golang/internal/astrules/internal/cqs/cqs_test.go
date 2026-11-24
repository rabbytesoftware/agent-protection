package cqs_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/cqs"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestCqs(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "returning a value while mutating the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (v int) {\n\tt.v = 5\n\treturn t.v\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "returning a value while writing through a pointer parameter is reported",
			code: "package p\n\nfunc f(\n\ttarget *int,\n) (v int) {\n\t*target = 5\n\treturn 5\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "pure query passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (v int) {\n\treturn t.v\n}\n",
			want: nil,
		},
		{
			name: "command returning only an error passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (err error) {\n\tt.v = 5\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "assigning to a value receiver copy passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t T) f() (v int) {\n\tt.v = 5\n\treturn t.v\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			cqs.Cqs([]source.File{parsed(
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
