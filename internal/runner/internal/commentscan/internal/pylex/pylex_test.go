package pylex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/pylex"
)

func TestToml(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "trailing comment is reported", code: "[a]\nb = 1 # why\n", want: []int{2}},
		{name: "hash in a string passes", code: "u = \"https://x.dev/#a\"\n", want: nil},
		{name: "multi-line basic string", code: "s = \"\"\"\n# not\n\"\"\"\n", want: nil},
		{name: "multi-line literal string", code: "s = '''\n# not\n'''\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			pylex.Toml(),
			tc.code,
			tc.want,
		)
	}
}

func TestGraphql(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment is reported", code: "type Query {\n  # deprecated\n  a: Int\n}\n", want: []int{2}},
		{name: "hash in a description passes", code: "\"\"\"\n# not a comment\n\"\"\"\ntype A { a: Int }\n", want: nil},
		{name: "hash in a string passes", code: "directive @a(x: String = \"#\") on FIELD\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			pylex.Graphql(),
			tc.code,
			tc.want,
		)
	}
}

func TestGraphqlMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		pylex.Graphql(),
		[]string{"a.graphql", "schema.graphqls", "q.gql"},
		[]string{"a.graph"},
	)
}

func TestPythonMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		pylex.Python(),
		[]string{"a.py", "b.pyi"},
		[]string{"a.pyc", "py"},
	)
}

func TestRubyMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		pylex.Ruby(),
		[]string{
			"a.rb",
			"Rakefile",
			"Gemfile",
			"x.gemspec",
			"t.rake",
		},
		[]string{"a.r", "Gemfile.lock"},
	)
}

func TestTomlMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		pylex.Toml(),
		[]string{"Cargo.toml", "a.toml"},
		[]string{"a.tml"},
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
