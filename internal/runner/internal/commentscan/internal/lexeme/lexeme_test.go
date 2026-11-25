package lexeme_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestLexemeOffsets(
	t *testing.T,
) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{name: "line end finds the newline", got: lexeme.LineEnd(
			[]byte("ab\ncd"),
			0,
		), want: 2},
		{name: "line end without a newline is the text end", got: lexeme.LineEnd(
			[]byte("ab\ncd"),
			3,
		), want: 5},
		{name: "shebang end is its line end", got: lexeme.ShebangEnd([]byte("#!/bin/sh\nx")), want: 9},
		{name: "no shebang ends at zero", got: lexeme.ShebangEnd([]byte("x")), want: 0},
		{name: "escape skips the escaped byte", got: lexeme.EscapedNext(
			[]byte("\\n"),
			0,
		), want: 2},
		{name: "plain byte advances by one", got: lexeme.EscapedNext(
			[]byte("a"),
			0,
		), want: 1},
		{name: "trailing escape stops at the text end", got: lexeme.EscapedNext(
			[]byte("\\"),
			0,
		), want: 1},
		{name: "quoted end skips an escaped quote", got: lexeme.QuotedEnd(
			[]byte("\"a\\\"b\" x"),
			0,
			lexeme.NoStops,
		), want: 6},
		{name: "quoted end stops at a line stop", got: lexeme.QuotedEnd(
			[]byte("\"ab\ncd"),
			0,
			lexeme.LineStops,
		), want: 3},
		{name: "raw end ignores escapes", got: lexeme.RawEnd(
			[]byte("'a\\'b"),
			0,
		), want: 4},
		{name: "unterminated raw end is the text end", got: lexeme.RawEnd(
			[]byte("'ab"),
			0,
		), want: 3},
		{name: "closed end past the text is clamped", got: lexeme.ClosedEnd(
			[]byte("ab"),
			5,
			'x',
		), want: 2},
		{name: "block end follows the closer", got: lexeme.BlockEnd(
			[]byte("/* a */ b"),
			0,
		), want: 7},
		{name: "unterminated block runs to the text end", got: lexeme.BlockEnd(
			[]byte("/* open"),
			0,
		), want: 7},
		{name: "next line starts after the newline", got: lexeme.NextLine(
			[]byte("ab\ncd"),
			0,
		), want: 3},
		{name: "next line on the last line is the text end", got: lexeme.NextLine(
			[]byte("ab\ncd"),
			3,
		), want: 5},
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

func TestLexemePredicates(
	t *testing.T,
) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "prefix at an offset matches", got: lexeme.HasPrefix(
			[]byte("a//b"),
			1,
			lexeme.LineCommentOpener,
		), want: true},
		{name: "prefix past the text does not match", got: lexeme.HasPrefix(
			[]byte("a/"),
			1,
			lexeme.LineCommentOpener,
		), want: false},
		{name: "letters are identifier bytes", got: lexeme.IsIdentifierByte('a'), want: true},
		{name: "dollar is an identifier byte", got: lexeme.IsIdentifierByte('$'), want: true},
		{name: "multibyte lead is an identifier byte", got: lexeme.IsIdentifierByte(0xC3), want: true},
		{name: "space is not an identifier byte", got: lexeme.IsIdentifierByte(' '), want: false},
		{name: "open cursor before a stop", got: lexeme.IsOpen(
			[]byte("ab"),
			0,
			lexeme.LineStops,
		), want: true},
		{name: "cursor at the text end is closed", got: lexeme.IsOpen(
			[]byte("ab"),
			2,
			lexeme.LineStops,
		), want: false},
	}
	for _, tc := range cases {
		assertPredicate(
			t,
			tc.name,
			tc.got,
			tc.want,
		)
	}
}

func TestLexemeScanUntil(
	t *testing.T,
) {
	got := lexeme.ScanUntil(
		[]byte(`a\;b;c`),
		0,
		";",
	)
	if got != 4 {
		t.Errorf(
			"ScanUntil() = %d, want 4 past the escaped stop",
			got,
		)
	}
}

func TestLexemeLineComment(
	t *testing.T,
) {
	end, spans := lexeme.LineComment(
		[]byte("a // b\nc"),
		2,
	)
	want := []lexeme.Span{{Start: 2, End: 6, IsLine: true}}
	isWrong := end != 6 || !slices.Equal(
		spans,
		want,
	)
	if isWrong {
		t.Errorf(
			"LineComment() = %d, %v, want 6, %v",
			end,
			spans,
			want,
		)
	}
}

func TestLexemeWithout(
	t *testing.T,
) {
	spans := []lexeme.Span{{Start: 0, End: 1}, {Start: 2, End: 3}}
	got := lexeme.Without(
		[]byte("a b"),
		spans,
		startsTheText,
	)
	if !slices.Equal(
		got,
		spans[1:],
	) {
		t.Errorf(
			"Without() = %v, want %v",
			got,
			spans[1:],
		)
	}
}

func startsTheText(
	_ []byte,
	span lexeme.Span,
) (directive bool) {
	return span.Start == 0
}

func assertOffset(
	t *testing.T,
	name string,
	got int,
	want int,
) {
	t.Helper()
	if got != want {
		t.Errorf(
			"%s: got %d, want %d",
			name,
			got,
			want,
		)
	}
}

func assertPredicate(
	t *testing.T,
	name string,
	got bool,
	want bool,
) {
	t.Helper()
	if got != want {
		t.Errorf(
			"%s: got %v, want %v",
			name,
			got,
			want,
		)
	}
}
