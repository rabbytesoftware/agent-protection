package stringcomment_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/stringcomment"
)

func TestStringcomment(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "blank const string is reported", path: "p.go", code: "package p\n\nconst _ = \"why\"\n", want: []string{"R07 p.go:3"}},
		{name: "blank raw string const is reported", path: "p.go", code: "package p\n\nconst _ = `\nwhy\n`\n", want: []string{"R07 p.go:3"}},
		{name: "blank var concatenation is reported", path: "p.go", code: "package p\n\nvar _ = \"a\" + (\"b\")\n", want: []string{"R07 p.go:3"}},
		{name: "blank assignment is reported", path: "p.go", code: "package p\n\nfunc F() {\n\t_ = \"HACK: skips rounding\"\n}\n", want: []string{"R07 p.go:4"}},
		{name: "blank pair assignment is reported", path: "p.go", code: "package p\n\nfunc F() {\n\t_, _ = \"a\", 1\n}\n", want: []string{"R07 p.go:4"}},
		{name: "blank var in a function is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tvar _ = \"x\"\n}\n", want: []string{"R07 p.go:4"}},
		{name: "interface assertion passes", path: "p.go", code: "package p\n\nvar _ error = (*e)(nil)\n", want: nil},
		{name: "named string const passes", path: "p.go", code: "package p\n\nconst Name = \"x\"\n", want: nil},
		{name: "blank beside a name passes", path: "p.go", code: "package p\n\nvar a, _ = \"x\", \"y\"\n", want: nil},
		{name: "blank number passes", path: "p.go", code: "package p\n\nconst _ = 1\n", want: nil},
		{name: "blank call passes", path: "p.go", code: "package p\n\nfunc F() {\n\t_ = len(\"x\")\n}\n", want: nil},
		{name: "string assigned to a named result passes", path: "p.go", code: "package p\n\nfunc F() (s string) {\n\ts = \"x\"\n\treturn s\n}\n", want: nil},
		{name: "string concatenated with a call passes", path: "p.go", code: "package p\n\nvar _ = \"a\" + f()\n", want: nil},
		{name: "string assigned to a field passes", path: "p.go", code: "package p\n\nfunc F() {\n\tx.y = \"a\"\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stringcomment.Stringcomment([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestStringcommentSnippet(
	t *testing.T,
) {
	findings := stringcomment.Stringcomment([]source.File{parsed(
		t,
		"p.go",
		"package p\n\nconst _ = \"why\"\n",
	)})
	want := []report.Finding{{Rule: "R07", Location: "p.go:3", Snippet: "string literal bound to _ is a comment moved into a string: delete it and make the code say it through names and structure"}}
	if !slices.Equal(
		findings,
		want,
	) {
		t.Errorf(
			"Stringcomment() = %v, want %v",
			findings,
			want,
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

	isWrong := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isWrong {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
