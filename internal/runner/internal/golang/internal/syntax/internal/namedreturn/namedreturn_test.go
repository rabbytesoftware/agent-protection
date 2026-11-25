package namedreturn_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/namedreturn"
)

func TestNamedreturn(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "naked return on named results is reported at the return",
			code: "package p\n\nfunc f() (result int, err error) {\n\treturn\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "explicit return on named results passes",
			code: "package p\n\nfunc f() (result int, err error) {\n\treturn 1, nil\n}\n",
			want: nil,
		},
		{
			name: "naked return in a literal with named results is reported",
			code: "package p\n\nvar f = func() (result int) {\n\treturn\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "bare return in a nested literal without results passes",
			code: "package p\n\nfunc f() (result int) {\n\tg := func() {\n\t\treturn\n\t}\n\tg()\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "var declared only to be returned is reported at its name",
			code: "package p\n\nfunc f() (count int) {\n\tvar total int\n\ttotal++\n\treturn total\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "var with an initializer returned by name is reported",
			code: "package p\n\nfunc f() (count int, err error) {\n\tvar total = 1\n\treturn total, nil\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "var used in a returned expression passes",
			code: "package p\n\nfunc f() (result int) {\n\tvar total int\n\ttotal++\n\treturn total + 1\n}\n",
			want: nil,
		},
		{
			name: "const returned by name passes",
			code: "package p\n\nfunc f() (result int) {\n\tconst total = 1\n\treturn total\n}\n",
			want: nil,
		},
		{
			name: "var returned by a nested literal is reported once",
			code: "package p\n\nfunc f() (result int) {\n\tg := func() (count int) {\n\t\tvar total int\n\t\treturn total\n\t}\n\treturn g()\n}\n",
			want: []string{"p.go:5"},
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			namedreturn.Namedreturn([]source.File{parsed(
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
