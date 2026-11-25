package jslex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/jslex"
	"testing"
)

func TestRegex(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "regex with escaped slashes passes", code: "const r = /https?:\\/\\//;\n", want: nil},
		{name: "division followed by a comment", code: "x = a / b // c\n", want: []int{1}},
		{name: "two divisions on one line pass", code: "x = a / b / c;\n", want: nil},
		{name: "regex after an open paren", code: "s.replace(/\\/\\/+/g, '/');\n", want: nil},
		{name: "regex after a comma", code: "f(a, /\\/\\*/);\n", want: nil},
		{name: "regex after a colon", code: "const o = { r: /\\/\\// };\n", want: nil},
		{name: "regex after an array opener", code: "const l = [/\\/\\//];\n", want: nil},
		{name: "regex after a negation", code: "if (!/\\/\\//.test(s)) {}\n", want: nil},
		{name: "regex after logical operators", code: "a && /\\/\\//.test(s) || /\\/\\*/.test(s);\n", want: nil},
		{name: "regex after a ternary", code: "x = c ? /\\/\\// : /\\/\\*/;\n", want: nil},
		{name: "regex after a semicolon", code: "a();/\\/\\//.test(s);\n", want: nil},
		{name: "regex after return", code: "function f() { return /\\/\\//; }\n", want: nil},
		{name: "regex after typeof", code: "x = typeof /\\/\\//;\n", want: nil},
		{name: "regex at line start", code: "a()\n  /\\/\\//.test(s)\n", want: nil},
		{name: "slash inside a character class does not end the regex", code: "const r = /[/]\\/\\//;\n", want: nil},
		{name: "character class with an escaped bracket", code: "const r = /[\\]/]// x/;\n", want: nil},
		{name: "quote inside a regex does not open a string", code: "const r = /'/; // real\n", want: []int{1}},
		{name: "identifier ending in return is division", code: "x = noreturn / 2 // c\n", want: []int{1}},
		{name: "unterminated regex stops at the line end", code: "x = (/oops\n// real\n", want: []int{2}},
		{name: "unterminated character class stops at the line end", code: "x = (/[ab\n// real\n", want: []int{2}},
		{name: "unterminated regex at end of file is safe", code: "x = (/ab", want: nil},
		{name: "unterminated character class at end of file is safe", code: "x = (/[ab", want: nil},
		{name: "division after a closing paren", code: "x = (a) / (b) // c\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			jslex.Jslex(),
			tc.code,
			tc.want,
		)
	}
}
