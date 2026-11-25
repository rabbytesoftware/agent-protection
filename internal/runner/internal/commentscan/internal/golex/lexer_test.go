package golex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/golex"
	"testing"
)

func TestSlashLexer(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "doc comment is reported", code: "package p\n\n// F does.\nfunc F() {}\n", want: []int{3}},
		{name: "string with slashes passes", code: "var s = \"// no\"\n", want: nil},
		{name: "raw string with slashes passes", code: "var s = `\n// no\n`\n", want: nil},
		{name: "rune literal passes", code: "var r = '/'\nvar s = \"// x\"\n", want: nil},
		{name: "unterminated raw string runs to the end", code: "var s = `open\n// in\n", want: nil},
		{name: "block comment is reported", code: "/* x */\npackage p\n", want: []int{1}},
		{name: "division passes", code: "x := a / b / c\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			golex.Gosource(),
			tc.code,
			tc.want,
		)
	}
}
