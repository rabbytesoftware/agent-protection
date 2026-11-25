package deadstore_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/deadstore"
)

func TestDeadstore(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "error never read before return is reported",
			code: "package p\n\nfunc f(\n\ts string,\n) (err error) {\n\t_, err = strconv.Atoi(s)\n\treturn nil\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "goto bails out",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\nagain:\n\terr = b()\n\tgoto again\n}\n",
			want: nil,
		},
		{
			name: "closure body is analyzed on its own",
			code: "package p\n\nvar g = func() (n int) {\n\tvalue, err := strconv.Atoi(\"1\")\n\treturn value\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "function without body passes",
			code: "package p\n\nfunc f() (err error)\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertCase(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func assertCase(
	t *testing.T,
	name string,
	code string,
	want []string,
) {
	t.Helper()
	assertLocations(
		t,
		name,
		deadstore.Deadstore([]source.File{parsed(
			t,
			code,
		)}),
		want,
	)
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
