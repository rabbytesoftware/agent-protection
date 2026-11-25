package sideeffect_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/sideeffect"
)

func TestSideeffect(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "query name that writes a file is reported",
			code: "package p\n\nimport \"os\"\n\nfunc GetConfig() (err error) {\n\treturn os.WriteFile(\"x\", nil, 0o644)\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "predicate method that deletes from a cache is reported",
			code: "package p\n\ntype S struct {\n\tcache map[string]int\n}\n\nfunc (s S) IsStale() (stale bool) {\n\ts.cache.Delete(\"k\")\n\treturn true\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "query name that only reads passes",
			code: "package p\n\nfunc GetValue() (value int) {\n\treturn 5\n}\n",
			want: nil,
		},
		{
			name: "query word cut short by the name is not a query",
			code: "package p\n\ntype S struct {\n\tv map[string]int\n}\n\nfunc (s S) Listen() {\n\ts.v.Store(1)\n}\n\nfunc (s S) Issue() {\n\ts.v.Send(1)\n}\n\nfunc (s S) Hashed() {\n\ts.v.Save(1)\n}\n",
			want: nil,
		},
		{
			name: "query word as the whole name or before a digit is reported",
			code: "package p\n\ntype S struct {\n\tv map[string]int\n}\n\nfunc (s S) List() {\n\ts.v.Store(1)\n}\n\nfunc (s S) Get2() {\n\ts.v.Store(1)\n}\n\nfunc (s S) CalculateTotal() {\n\ts.v.Store(1)\n}\n",
			want: []string{"p.go:7", "p.go:11", "p.go:15"},
		},
		{
			name: "honest command name that writes passes",
			code: "package p\n\nimport \"os\"\n\nfunc SaveConfig() (err error) {\n\treturn os.WriteFile(\"x\", nil, 0o644)\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			sideeffect.Sideeffect([]source.File{parsed(
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
