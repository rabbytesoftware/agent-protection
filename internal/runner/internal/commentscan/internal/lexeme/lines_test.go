package lexeme_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestLines(
	t *testing.T,
) {
	cases := []struct {
		name  string
		text  string
		spans []lexeme.Span
		want  []int
	}{
		{name: "adjacent line comments are one run", text: "// one\n// two\na\n// three\n", spans: []lexeme.Span{{Start: 0, End: 6, IsLine: true}, {Start: 7, End: 13, IsLine: true}, {Start: 16, End: 24, IsLine: true}}, want: []int{1, 4}},
		{name: "a blank line splits a run", text: "// a\n\n// b\n", spans: []lexeme.Span{{Start: 0, End: 4, IsLine: true}, {Start: 6, End: 10, IsLine: true}}, want: []int{1, 3}},
		{name: "block comments are never merged", text: "/* a */\n/* b */\n", spans: []lexeme.Span{{Start: 0, End: 7}, {Start: 8, End: 15}}, want: []int{1, 2}},
		{name: "no spans give no lines", text: "a\n", spans: nil, want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			lexeme.Lines(
				[]byte(tc.text),
				tc.spans,
			),
			tc.want,
		)
	}
}

func assertLines(
	t *testing.T,
	name string,
	got []int,
	want []int,
) {
	t.Helper()
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: Lines() = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func TestLinePositions(
	t *testing.T,
) {
	text := []byte("ab\n  # c\nd # e")
	cases := []struct {
		name string
		got  int
		want int
	}{
		{name: "line of an offset on the third line", got: lexeme.LineOf(
			text,
			12,
		), want: 3},
		{name: "line start of an offset on the second line", got: lexeme.LineStart(
			text,
			6,
		), want: 3},
		{name: "line start on the first line is zero", got: lexeme.LineStart(
			text,
			1,
		), want: 0},
	}
	for _, tc := range cases {
		assertOffset(
			t,
			tc.name,
			tc.got,
			tc.want,
		)
	}
}

func TestIsLineLeading(
	t *testing.T,
) {
	text := []byte("ab\n  # c\nd # e")
	isLeading := lexeme.IsLineLeading(
		text,
		lexeme.Span{Start: 5, End: 8},
	)
	isTrailing := !lexeme.IsLineLeading(
		text,
		lexeme.Span{Start: 11, End: 14},
	)
	isWrong := !isLeading || !isTrailing
	if isWrong {
		t.Errorf(
			"IsLineLeading() leading = %v, trailing = %v, want true and true",
			isLeading,
			isTrailing,
		)
	}
}
