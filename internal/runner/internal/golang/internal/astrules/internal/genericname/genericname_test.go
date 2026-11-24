package genericname_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/genericname"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestGenericname(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "function built from role words is reported",
			code: "package p\n\nfunc runHandler() {}\n",
			want: []string{"R56 p.go:3"},
		},
		{
			name: "lone filler word is reported",
			code: "package p\n\ntype helper struct{}\n",
			want: []string{"R56 p.go:3"},
		},
		{
			name: "lone role word names its role and passes",
			code: "package p\n\ntype handler struct{}\n\nfunc run() {}\n",
			want: nil,
		},
		{
			name: "role word with a meaningful word passes",
			code: "package p\n\nfunc handleDispatch() {}\n\nfunc makeEventBlob() {}\n",
			want: nil,
		},
		{
			name: "content-free prefix is reported",
			code: "package p\n\nfunc implSendCreate() {}\n",
			want: []string{"R56 p.go:3"},
		},
		{
			name: "filler word alone with a number is reported",
			code: "package p\n\nfunc setup() {\n\tfn2 := func() {}\n\t_ = fn2\n}\n",
			want: []string{"R56 p.go:4"},
		},
		{
			name: "closure numbered beside its sibling is reported",
			code: "package p\n\nfunc setup() {\n\tack := func() {}\n\tack2 := func() {}\n\t_, _ = ack, ack2\n}\n",
			want: []string{"R56 p.go:5"},
		},
		{
			name: "numbered siblings without the bare name are all reported",
			code: "package p\n\nfunc match2() {}\n\nfunc match3() {}\n",
			want: []string{"R56 p.go:3", "R56 p.go:5"},
		},
		{
			name: "number without a sibling passes",
			code: "package p\n\ntype v0 struct{}\n\nfunc toBase64() {}\n",
			want: nil,
		},
		{
			name: "number after a capital is part of the name",
			code: "package p\n\nfunc latencyP50() {}\n\nfunc latencyP99() {}\n\nfunc deliveryA1() {}\n\nfunc deliveryA2() {}\n",
			want: nil,
		},
		{
			name: "var closure is checked",
			code: "package p\n\nvar wrap = func() {}\n",
			want: []string{"R56 p.go:3"},
		},
		{
			name: "non-closure values and field targets are ignored",
			code: "package p\n\nvar data = 1\n\nfunc setup(s *struct{ fn func() }) {\n\ts.fn = func() {}\n\tdo := 2\n\t_ = do\n}\n",
			want: nil,
		},
		{
			name: "unexported method is checked, exported names are not",
			code: "package p\n\ntype t struct{}\n\nfunc (t) work() {}\n\nfunc (t) Handle() {}\n\nfunc RunHandler() {}\n",
			want: []string{"R56 p.go:5"},
		},
		{
			name: "program entry, package setup and blank are exempt",
			code: "package main\n\nfunc main() {}\n\nfunc init() {}\n\nvar _ = func() {}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			genericname.Genericname([]source.File{parsed(
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
