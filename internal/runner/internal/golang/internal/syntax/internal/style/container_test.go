package style_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style"
)

func TestStyle(
	t *testing.T,
) {
	file, err := source.Parse(
		".",
		"p.go",
		[]byte("package p\n\n// Pick picks.\nfunc Pick(x int) (label string) {\n\tif x > 0 {\n\t\treturn \"a\"\n\t} else {\n\t\treturn \"b\"\n\t}\n}\n"),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	findings := style.Style(
		[]source.File{file},
	)
	rules := []string{}
	for _, finding := range findings {
		rules = append(
			rules,
			finding.Rule,
		)
	}

	for _, want := range []string{"R01", "R03"} {
		assertContains(
			t,
			rules,
			want,
		)
	}
}

func assertContains(
	t *testing.T,
	rules []string,
	want string,
) {
	t.Helper()
	if !slices.Contains(
		rules,
		want,
	) {
		t.Errorf(
			"style rules report comments and else: rules = %v, want %s",
			rules,
			want,
		)
	}
}
