package astrules_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestAstrules(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "source files get layout and design rules",
			code: "package p\n\n// Pick picks.\nfunc Pick(x bool) (label string) {\n\tif x {\n\t\treturn \"a\"\n\t} else {\n\t\treturn \"b\"\n\t}\n}\n",
			want: []string{"R01", "R03", "R32"},
		},
	}
	for _, tc := range cases {
		findings := astrules.Astrules(
			[]source.File{parsed(
				t,
				tc.code,
			)},
			nil,
		)
		assertRules(
			t,
			tc.name,
			findings,
			tc.want,
		)
	}
}

func TestAstrulesAppliesLayoutRulesToTests(
	t *testing.T,
) {
	test := parsed(
		t,
		"package p\n\nfunc TestX(t *testing.T) {\n\tif true {\n\t} else {\n\t}\n}\n",
	)
	findings := astrules.Astrules(
		nil,
		[]source.File{test},
	)
	assertRules(
		t,
		"test files get layout rules",
		findings,
		[]string{"R01", "R03"},
	)
}

func assertRules(
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
			finding.Rule,
		)
	}

	missing := []string{}
	for _, rule := range want {
		missing = append(
			missing,
			absent(
				got,
				rule,
			)...,
		)
	}

	if len(missing) > 0 {
		t.Errorf(
			"%s: rules %v missing from %v",
			name,
			missing,
			got,
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

func absent(
	got []string,
	rule string,
) (missing []string) {
	if slices.Contains(
		got,
		rule,
	) {
		return nil
	}

	return []string{rule}
}
