package linelex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/linelex"
)

func TestIgnore(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "leading hash is reported", code: "# generated junk\n*.tmp\n", want: []int{1}},
		{name: "hash inside a pattern passes", code: "foo#bar\n", want: nil},
		{name: "indented hash is reported", code: "  # why\n", want: []int{1}},
		{name: "comment run is one finding", code: "# a\n# b\nx\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			linelex.Ignore(),
			tc.code,
			tc.want,
		)
	}
}
func TestIni(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "semicolon comment is reported", code: "[core]\n; legacy setting\nx=1\n", want: []int{2}},
		{name: "hash comment is reported", code: "# why\n[core]\n", want: []int{1}},
		{name: "values pass", code: "[core]\nx=1;2\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			linelex.Ini(),
			tc.code,
			tc.want,
		)
	}
}
func TestProperties(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "bang comment is reported", code: "! why\na=1\n", want: []int{1}},
		{name: "hash comment is reported", code: "# why\na=1\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			linelex.Properties(),
			tc.code,
			tc.want,
		)
	}
}
func TestFallback(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "hash line is reported", code: "a: 1\n# why\n", want: []int{2}},
		{name: "slash line is reported", code: "// why\n", want: []int{1}},
		{name: "dash line is reported", code: "-- why\n", want: []int{1}},
		{name: "bare dash line is reported", code: "--\n", want: []int{1}},
		{name: "semicolon line is reported", code: "; why\n", want: []int{1}},
		{name: "block opener is reported", code: "/* why */\n", want: []int{1}},
		{name: "markup comment is reported", code: "<!-- why -->\n", want: []int{1}},
		{name: "shebang passes", code: "#!/usr/bin/env perl\nprint 1;\n", want: nil},
		{name: "comment after a shebang is reported", code: "#!/usr/bin/env perl\n# why\n", want: []int{2}},
		{name: "flags and rules pass", code: "--flag=1\n---\n-----BEGIN CERTIFICATE-----\n", want: nil},
		{name: "code lines pass", code: "a = 1 # trailing\n", want: nil},
		{name: "windows line endings are trimmed", code: "x\r\n# why\r\n", want: []int{2}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			linelex.Fallback(),
			tc.code,
			tc.want,
		)
	}
}

func TestIgnoreMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		linelex.Ignore(),
		[]string{
			".gitignore",
			".dockerignore",
			".gitattributes",
			"CODEOWNERS",
		},
		[]string{"gitignore", "a.ignore"},
	)
}

func TestIniMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		linelex.Ini(),
		[]string{
			"a.ini",
			"setup.cfg",
			"nginx.conf",
			".editorconfig",
			".npmrc",
		},
		[]string{"a.inix"},
	)
}

func TestPropertiesMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		linelex.Properties(),
		[]string{"a.properties"},
		[]string{"a.props"},
	)
}

func TestFallbackMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		linelex.Fallback(),
		[]string{},
		[]string{"anything", "a.txt"},
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
