package behavior_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior"
)

func TestBehavior(
	t *testing.T,
) {
	file, err := source.Parse(
		".",
		"p.go",
		[]byte("package p\n\nfunc Pick(x bool) (label string) {\n\tif x {\n\t\treturn \"a\"\n\t}\n\n\treturn \"b\"\n}\n"),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	findings := behavior.Behavior(
		[]source.File{file},
		nil,
	)
	rules := []string{}
	for _, finding := range findings {
		rules = append(
			rules,
			finding.Rule,
		)
	}

	for _, want := range []string{"R32"} {
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
			"behavior rules report bool parameters: rules = %v, want %s",
			rules,
			want,
		)
	}
}
