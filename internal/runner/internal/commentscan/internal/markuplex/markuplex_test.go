package markuplex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/markuplex"
)

func TestMarkup(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "markup comment is reported", code: "<!-- TODO: remove -->\n<p>x</p>\n", want: []int{1}},
		{name: "template comment is reported", code: "<p>{{/* hack */}}{{.Total}}</p>\n", want: []int{1}},
		{name: "trimmed template comment is reported", code: "{{- /* why */ -}}\n", want: []int{1}},
		{name: "template comment holding braces is one span", code: "{{/* a }} b */}}\n<p/>\n", want: []int{1}},
		{name: "template action passes", code: "<p>{{.Total}}</p>\n", want: nil},
		{name: "doctype and declaration pass", code: "<?xml version=\"1.0\"?>\n<!DOCTYPE html>\n", want: nil},
		{name: "conditional comment is reported", code: "<!--[if IE]><p/><![endif]-->\n", want: []int{1}},
		{name: "cdata passes", code: "<a><![CDATA[ <!-- not --> ]]></a>\n", want: nil},
		{name: "script body is lexed by the script language", code: "<p/>\n<script>x\n</script>\n", want: []int{2}},
		{name: "uppercase script tag is lexed", code: "<SCRIPT type=\"m\">x</SCRIPT>\n", want: []int{1}},
		{name: "style body is lexed by the style language", code: "<style>\nx\n</style>\n", want: nil},
		{name: "unclosed script runs to the end", code: "<script>x\n", want: []int{1}},
		{name: "a tag that only starts like script passes", code: "<scripted>x</scripted>\n", want: nil},
		{name: "unterminated comment runs to the end", code: "<!-- open\n<p/>\n", want: []int{1}},
		{name: "unterminated action runs to the end", code: "{{ .A\n", want: nil},
		{name: "lone angle bracket at the end is safe", code: "<", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			markuplex.Markup(
				lexeme.Language{Spans: wholeBody},
				lexeme.Language{Spans: noSpans},
			),
			tc.code,
			tc.want,
		)
	}
}

func TestMarkupMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		markuplex.Markup(
			lexeme.Language{},
			lexeme.Language{},
		),
		[]string{
			"a.html",
			"b.HTM",
			"c.tmpl",
			"d.gohtml",
			"e.xml",
			"f.svg",
		},
		[]string{"a.htmlx", "html"},
	)
}

func wholeBody(
	text []byte,
) (spans []lexeme.Span) {
	if len(text) == 0 {
		return nil
	}

	return []lexeme.Span{{Start: 0, End: len(text)}}
}

func noSpans(
	_ []byte,
) (spans []lexeme.Span) {
	return nil
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
