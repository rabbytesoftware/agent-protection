package returnperline_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/returnperline"
)

func TestReturnperline(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "bare and single-value returns pass", code: "package p\n\nfunc f() {\n\treturn\n}\n\nfunc g() int {\n\treturn 1\n}\n", want: nil},
		{name: "two values on one line are reported", code: "package p\n\nfunc f() (int, error) {\n\treturn 1, nil\n}\n", want: []string{"R54 p.go:4"}},
		{name: "one value per line passes", code: "package p\n\nfunc f() (int, error) {\n\treturn 1,\n\t\tnil\n}\n", want: nil},
		{name: "a later pair sharing a line is reported", code: "package p\n\nfunc f() (int, int, error) {\n\treturn 1,\n\t\t2, nil\n}\n", want: []string{"R54 p.go:4"}},
		{name: "a multi-line value followed by a value on its closing line is reported", code: "package p\n\nfunc f() (error, int) {\n\treturn g(\n\t\t1,\n\t), 2\n}\n", want: []string{"R54 p.go:4"}},
		{name: "returns inside func literals are checked", code: "package p\n\nvar f = func() (int, error) {\n\treturn 1, nil\n}\n", want: []string{"R54 p.go:4"}},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			returnperline.Returnperline([]source.File{parsed(
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
	) &&
		len(got)+len(want) > 0
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

	return source.File{
		Path:    "p.go",
		Rel:     "p.go",
		FileSet: fileSet,
		Syntax:  syntax,
	}
}
