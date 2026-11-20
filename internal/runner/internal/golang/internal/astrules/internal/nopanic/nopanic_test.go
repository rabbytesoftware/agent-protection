package nopanic_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/nopanic"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestNopanic(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "panic in a library package is reported at the call",
			code: "package p\n\nfunc f() {\n\tpanic(\"boom\")\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "panic inside a function literal is reported",
			code: "package p\n\nvar f = func() {\n\tpanic(\"boom\")\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "panic in package main passes",
			code: "package main\n\nfunc main() {\n\tpanic(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "method named panic passes",
			code: "package p\n\nfunc f(\n\tx alarm,\n) {\n\tx.panic()\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			nopanic.Nopanic([]source.File{parsed(
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
