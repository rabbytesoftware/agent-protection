package paragraph_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/paragraph"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestParagraph(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "statement right after a block is reported", code: "package p\n\nfunc f() {\n\tif a {\n\t\tg()\n\t}\n\tg()\n}\n", want: []string{"R55 p.go:7"}},
		{name: "blank line after a block passes", code: "package p\n\nfunc f() {\n\tif a {\n\t\tg()\n\t}\n\n\tg()\n}\n", want: nil},
		{name: "block closing its enclosing block passes", code: "package p\n\nfunc f() {\n\tg()\n\tfor range a {\n\t\tg()\n\t}\n}\n", want: nil},
		{name: "loops, switches, selects and bare blocks all count", code: "package p\n\nfunc f() {\n\tfor range a {\n\t}\n\tswitch {\n\t}\n\tselect {\n\t}\n\t{\n\t}\n\tg()\n}\n", want: []string{"R55 p.go:6", "R55 p.go:8", "R55 p.go:10", "R55 p.go:12"}},
		{name: "labeled loop counts as a block", code: "package p\n\nfunc f() {\nouter:\n\tfor range a {\n\t\tbreak outer\n\t}\n\tg()\n}\n", want: []string{"R55 p.go:8"}},
		{name: "case and comm clause bodies are checked", code: "package p\n\nfunc f() {\n\tswitch {\n\tcase a:\n\t\tif b {\n\t\t}\n\t\tg()\n\t}\n\tselect {\n\tcase <-c:\n\t\tif b {\n\t\t}\n\t\tg()\n\t}\n}\n", want: []string{"R55 p.go:10", "R55 p.go:8", "R55 p.go:14"}},
		{name: "check detached from its assignment is reported", code: "package p\n\nfunc f() {\n\terr := g()\n\n\tif err != nil {\n\t}\n}\n", want: []string{"R55 p.go:6"}},
		{name: "check attached to its assignment passes", code: "package p\n\nfunc f() {\n\terr := g()\n\tif err != nil {\n\t}\n}\n", want: nil},
		{name: "blank before an unrelated if passes", code: "package p\n\nfunc f() {\n\tx := g()\n\n\tif a {\n\t}\n\t_ = x\n}\n", want: []string{"R55 p.go:8"}},
		{name: "if with its own init is not a detached check", code: "package p\n\nfunc f() {\n\terr := g()\n\n\tif err := g(); err != nil {\n\t}\n}\n", want: nil},
		{name: "leading blank line in a block is reported", code: "package p\n\nfunc f() {\n\n\tg()\n}\n", want: []string{"R55 p.go:3"}},
		{name: "trailing blank line in a block is reported", code: "package p\n\nfunc f() {\n\tg()\n\n}\n", want: []string{"R55 p.go:3"}},
		{name: "empty blocks pass", code: "package p\n\nfunc f() {\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			paragraph.Paragraph([]source.File{parsed(
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
