package argumentperline_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/argumentperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestArgumentperline(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "zero and one argument stay inline", code: "package p\n\nfunc f() {\n\tg()\n\th(1)\n}\n", want: nil},
		{name: "two arguments one per line pass", code: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,\n\t)\n}\n", want: nil},
		{name: "two inline arguments are reported", code: "package p\n\nfunc f() {\n\tg(1, 2)\n}\n", want: []string{"R52 p.go:4"}},
		{name: "first argument on the paren line is reported", code: "package p\n\nfunc f() {\n\tg(1,\n\t\t2,\n\t)\n}\n", want: []string{"R52 p.go:4"}},
		{name: "closing paren on the last argument line is reported", code: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2)\n}\n", want: []string{"R52 p.go:4"}},
		{name: "nested inline call is reported inside a compliant outer call", code: "package p\n\nfunc f() {\n\tg(\n\t\th(1, 2),\n\t\t3,\n\t)\n}\n", want: []string{"R52 p.go:5"}},
		{name: "variadic spread with one argument per line passes", code: "package p\n\nfunc f() {\n\tx = append(\n\t\tx,\n\t\ty...,\n\t)\n}\n", want: nil},
		{name: "multi-line argument sharing a line with the next is reported", code: "package p\n\nfunc f() {\n\tg(\n\t\th(\n\t\t\t1,\n\t\t), 2,\n\t)\n}\n", want: []string{"R52 p.go:4"}},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			argumentperline.Argumentperline([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
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
