package paramperline_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/paramperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestParamperline(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "one parameter per line passes",
			code: `package p

func f(
	a int,
	b string,
) {
}
`,
			want: nil,
		},
		{
			name: "empty parameter list passes",
			code: `package p

func f() {
}
`,
			want: nil,
		},
		{
			name: "single inline parameter is reported",
			code: `package p

func f(a int) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "two names sharing a type on the signature line are reported",
			code: `package p

func f(a, b int) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "two names sharing a type on their own line are reported",
			code: `package p

func f(
	a, b int,
) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "two fields on one line are reported",
			code: `package p

func f(
	a int, b string,
) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "first parameter on the opening paren line is reported",
			code: `package p

func f(a int,
	b string,
) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "closing paren on the last parameter line is reported",
			code: `package p

func f(
	a int,
	b string) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "parameter starting on the line a multi-line parameter ends is reported",
			code: `package p

func f(
	callback func(
		x int,
	) error, b int,
) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "inline receiver with one parameter per line passes",
			code: `package p

type t struct{}

func (r t) m(
	a int,
) {
}
`,
			want: nil,
		},
		{
			name: "method with an inline parameter is reported",
			code: `package p

type t struct{}

func (r t) m(a int) {
}
`,
			want: []string{"R01 p.go:5"},
		},
		{
			name: "results on one line are not checked",
			code: `package p

func f(
	a int,
) (b int, c string) {
	return a, ""
}
`,
			want: nil,
		},
		{
			name: "type parameters on one line are not checked",
			code: `package p

func f[K comparable, V any](
	key K,
	value V,
) {
}
`,
			want: nil,
		},
		{
			name: "func literal with an inline parameter is reported",
			code: `package p

var f = func(a int) {
}
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "func literal with one parameter per line passes",
			code: `package p

var f = func(
	a int,
) {
}
`,
			want: nil,
		},
		{
			name: "func literal passed as an argument is reported",
			code: `package p

func f(
	run func(
		x int,
	),
) {
	run(1)
}

func g() {
	f(func(x int) {})
}
`,
			want: []string{"R01 p.go:12"},
		},
		{
			name: "interface method with an inline parameter is reported",
			code: `package p

type i interface {
	m(a int) error
	n() error
}
`,
			want: []string{"R01 p.go:4"},
		},
		{
			name: "interface method with one parameter per line passes",
			code: `package p

type i interface {
	m(
		a int,
	) error
}
`,
			want: nil,
		},
		{
			name: "func type declaration with inline parameters is reported",
			code: `package p

type handler func(a int, b string) error
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "func type declaration with unnamed inline parameters is reported",
			code: `package p

type handler func(int, string)
`,
			want: []string{"R01 p.go:3"},
		},
		{
			name: "func type declaration with one parameter per line passes",
			code: `package p

type handler func(
	int,
	string,
)
`,
			want: nil,
		},
		{
			name: "func typed parameter and struct field are not declarations",
			code: `package p

type s struct {
	callback func(a int, b string)
}

func f(
	callback func(int) error,
) {
}
`,
			want: nil,
		},
		{
			name: "every offending declaration is reported in source order",
			code: `package p

func f(a int) {
}

func g(
	a int,
) {
}

func h(a int) {
}
`,
			want: []string{"R01 p.go:3", "R01 p.go:11"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			paramperline.Paramperline([]source.File{parsed(
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
