package shape_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape"
)

func TestShape(
	t *testing.T,
) {
	file, err := source.Parse(
		".",
		"p.go",
		[]byte("package p\n\ntype A struct{ n int }\n\ntype B struct{ n int }\n\nconst unused = 1\n"),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	findings := shape.Shape(
		[]source.File{file},
		[]source.File{file},
	)
	rules := []string{}
	for _, finding := range findings {
		rules = append(
			rules,
			finding.Rule,
		)
	}

	for _, want := range []string{"R08", "R69"} {
		assertContains(
			t,
			rules,
			want,
		)
	}
}

func TestShapeCapsTestFileLength(
	t *testing.T,
) {
	long, err := source.Parse(
		".",
		"p_test.go",
		[]byte("package p\n"+strings.Repeat(
			"\n",
			600,
		)),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	findings := shape.Shape(
		nil,
		[]source.File{long},
	)
	rules := []string{}
	for _, finding := range findings {
		rules = append(
			rules,
			finding.Rule,
		)
	}

	assertContains(
		t,
		rules,
		"R44",
	)
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
			"shape rules report two types in one file: rules = %v, want %s",
			rules,
			want,
		)
	}
}
