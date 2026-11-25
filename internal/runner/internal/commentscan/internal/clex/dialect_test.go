package clex_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/clex"
)

func TestDialect(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want int
	}{
		{name: "unterminated string runs to the end of its line", code: "a = \"x\n// real\n", want: 1},
		{name: "unterminated block runs to the end", code: "/* open\nmessage A {}\n", want: 1},
		{name: "scheme separator is not a comment", code: "a = url(http://x)\n", want: 0},
		{name: "empty text has no comments", code: "", want: 0},
	}
	for _, tc := range cases {
		assertSpanCount(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func assertSpanCount(
	t *testing.T,
	name string,
	code string,
	want int,
) {
	t.Helper()
	got := len(clex.Proto().Spans([]byte(code)))
	if got != want {
		t.Errorf(
			"%s: spans = %d, want %d",
			name,
			got,
			want,
		)
	}
}
