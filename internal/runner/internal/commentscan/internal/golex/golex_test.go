package golex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/golex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestGomodMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		golex.Gomod(),
		[]string{"go.mod", "go.work"},
		[]string{"go.sum", "a.mod"},
	)
}

func TestGosourceMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		golex.Gosource(),
		[]string{"a.go", "x/y_test.go"},
		[]string{"a.goo", "go"},
	)
}

func assertLines(
	t *testing.T,
	name string,
	language lexeme.Language,
	code string,
	want []int,
) {
	t.Helper()
	got := lexeme.Lines(
		[]byte(code),
		language.Spans([]byte(code)),
	)
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: comment lines = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func assertMatches(
	t *testing.T,
	language lexeme.Language,
	matching []string,
	other []string,
) {
	t.Helper()
	got := []string{}
	for _, name := range slices.Concat(
		matching,
		other,
	) {
		got = append(
			got,
			matched(
				language,
				name,
			)...,
		)
	}

	if !slices.Equal(
		got,
		matching,
	) {
		t.Errorf(
			"matched names = %v, want %v",
			got,
			matching,
		)
	}
}

func matched(
	language lexeme.Language,
	name string,
) (names []string) {
	if !language.Matches(name) {
		return nil
	}

	return []string{name}
}
