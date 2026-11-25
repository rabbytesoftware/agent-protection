package linelex_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/linelex"
)

func TestMarkers(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want int
	}{
		{name: "last line without a newline is read", code: "x\n# why", want: 1},
		{name: "empty lines pass", code: "\n\n", want: 0},
		{name: "empty text passes", code: "", want: 0},
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
	got := len(linelex.Ignore().Spans([]byte(code)))
	if got != want {
		t.Errorf(
			"%s: spans = %d, want %d",
			name,
			got,
			want,
		)
	}
}
