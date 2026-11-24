package packagehygiene_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/packagehygiene"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestPackagehygiene(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "generic util package is reported",
			code: "package util\n\nfunc A() {}\n",
			want: []string{"R22 p.go:1"},
		},
		{
			name: "generic shared package is reported",
			code: "package shared\n\nfunc A() {}\n",
			want: []string{"R22 p.go:1"},
		},
		{
			name: "underscored package name is reported",
			code: "package my_pkg\n\nfunc A() {}\n",
			want: []string{"R21 p.go:1"},
		},
		{
			name: "mixed case package name is reported",
			code: "package myPkg\n\nfunc A() {}\n",
			want: []string{"R21 p.go:1"},
		},
		{
			name: "short lowercase noun passes",
			code: "package domain\n\nfunc A() {}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagehygiene.Packagehygiene([]source.File{parsed(
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
