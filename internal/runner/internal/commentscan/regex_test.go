package commentscan_test

import "testing"

func TestRegex(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "regex with escaped slashes passes", path: "a.ts", code: "const r = /https?:\\/\\//;\n", want: nil},
		{name: "division followed by a comment", path: "a.ts", code: "x = a / b // c\n", want: []string{"R07 a.ts:1"}},
		{name: "two divisions on one line pass", path: "a.ts", code: "x = a / b / c;\n", want: nil},
		{name: "regex after an open paren", path: "a.ts", code: "s.replace(/\\/\\/+/g, '/');\n", want: nil},
		{name: "regex after a comma", path: "a.ts", code: "f(a, /\\/\\*/);\n", want: nil},
		{name: "regex after a colon", path: "a.ts", code: "const o = { r: /\\/\\// };\n", want: nil},
		{name: "regex after an array opener", path: "a.ts", code: "const l = [/\\/\\//];\n", want: nil},
		{name: "regex after a negation", path: "a.ts", code: "if (!/\\/\\//.test(s)) {}\n", want: nil},
		{name: "regex after logical operators", path: "a.ts", code: "a && /\\/\\//.test(s) || /\\/\\*/.test(s);\n", want: nil},
		{name: "regex after a ternary", path: "a.ts", code: "x = c ? /\\/\\// : /\\/\\*/;\n", want: nil},
		{name: "regex after a semicolon", path: "a.ts", code: "a();/\\/\\//.test(s);\n", want: nil},
		{name: "regex after return", path: "a.ts", code: "function f() { return /\\/\\//; }\n", want: nil},
		{name: "regex after typeof", path: "a.ts", code: "x = typeof /\\/\\//;\n", want: nil},
		{name: "regex at line start", path: "a.ts", code: "a()\n  /\\/\\//.test(s)\n", want: nil},
		{name: "slash inside a character class does not end the regex", path: "a.ts", code: "const r = /[/]\\/\\//;\n", want: nil},
		{name: "character class with an escaped bracket", path: "a.ts", code: "const r = /[\\]/]// x/;\n", want: nil},
		{name: "quote inside a regex does not open a string", path: "a.ts", code: "const r = /'/; // real\n", want: []string{"R07 a.ts:1"}},
		{name: "identifier ending in return is division", path: "a.ts", code: "x = noreturn / 2 // c\n", want: []string{"R07 a.ts:1"}},
		{name: "unterminated regex stops at the line end", path: "a.ts", code: "x = (/oops\n// real\n", want: []string{"R07 a.ts:2"}},
		{name: "unterminated character class stops at the line end", path: "a.ts", code: "x = (/[ab\n// real\n", want: []string{"R07 a.ts:2"}},
		{name: "unterminated regex at end of file is safe", path: "a.ts", code: "x = (/ab", want: nil},
		{name: "unterminated character class at end of file is safe", path: "a.ts", code: "x = (/[ab", want: nil},
		{name: "division after a closing paren", path: "a.ts", code: "x = (a) / (b) // c\n", want: []string{"R07 a.ts:1"}},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
