package conditionperline_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/conditionperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestConditionperline(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "single condition passes", code: "package p\n\nfunc f() {\n\tif a {\n\t}\n}\n", want: nil},
		{name: "comparison is a single condition", code: "package p\n\nfunc f() {\n\tif a == b {\n\t}\n}\n", want: nil},
		{name: "joined conditions on one line are reported", code: "package p\n\nfunc f() {\n\tif a && b {\n\t}\n}\n", want: []string{"R53 p.go:4"}},
		{name: "or-joined conditions on one line are reported", code: "package p\n\nfunc f() {\n\tif a || b {\n\t}\n}\n", want: []string{"R53 p.go:4"}},
		{name: "one condition per line passes", code: "package p\n\nfunc f() {\n\tif a &&\n\t\tb {\n\t}\n}\n", want: nil},
		{name: "nested group must also split", code: "package p\n\nfunc f() {\n\tif a &&\n\t\t(b || c) {\n\t}\n}\n", want: []string{"R53 p.go:4"}},
		{name: "fully split nested group passes", code: "package p\n\nfunc f() {\n\tif a &&\n\t\t(b ||\n\t\t\tc) {\n\t}\n}\n", want: nil},
		{name: "joined booleans outside an if are not checked", code: "package p\n\nfunc f() {\n\tok := a && b\n\t_ = ok\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			conditionperline.Conditionperline([]source.File{parsed(
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
