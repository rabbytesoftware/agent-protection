package comments_test

import "testing"

func TestLexeme(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "url in a double quoted string passes", path: "a.ts", code: "const u = \"https://x.dev\";\n", want: nil},
		{name: "comment opener in a single quoted string passes", path: "a.ts", code: "const u = '/* not */ // not';\n", want: nil},
		{name: "escaped quote keeps the string open", path: "a.ts", code: "const u = \"a\\\" // still string\";\n", want: nil},
		{name: "escaped backslash closes the string", path: "a.ts", code: "const u = \"a\\\\\"; // real\n", want: []string{"R07 a.ts:1"}},
		{name: "unterminated ts string stops at the line end", path: "a.ts", code: "const u = 'oops\n// real\n", want: []string{"R07 a.ts:2"}},
		{name: "shebang on line one passes", path: "a.js", code: "#!/usr/bin/env node\nrun();\n", want: nil},
		{name: "shebang with a double slash passes", path: "a.mjs", code: "#!/usr/bin/env -S node --x=//y\nrun();\n", want: nil},
		{name: "unterminated block comment is reported", path: "a.ts", code: "a();\n/* open\n", want: []string{"R07 a.ts:2"}},
		{name: "trailing backslash at end of file is safe", path: "a.ts", code: "const u = '\\", want: nil},
		{name: "comment at end of file without newline is reported", path: "a.ts", code: "a(); // end", want: []string{"R07 a.ts:1"}},
		{name: "unicode identifiers do not confuse the lexer", path: "a.ts", code: "const café = 'x'; // c\n", want: []string{"R07 a.ts:1"}},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
