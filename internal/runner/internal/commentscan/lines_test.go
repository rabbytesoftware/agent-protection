package commentscan_test

import "testing"

func TestLines(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "consecutive line comments are one finding", path: "a.ts", code: "// one\n// two\n  // three\nconst a = 1;\n", want: []string{"R07 a.ts:1"}},
		{name: "a blank line splits a run", path: "a.ts", code: "// one\n\n// two\n", want: []string{"R07 a.ts:1", "R07 a.ts:3"}},
		{name: "code between line comments splits a run", path: "a.ts", code: "// one\nconst a = 1;\n// two\n", want: []string{"R07 a.ts:1", "R07 a.ts:3"}},
		{name: "trailing comments on code lines are separate", path: "a.ts", code: "a(); // one\nb(); // two\n", want: []string{"R07 a.ts:1", "R07 a.ts:2"}},
		{name: "block comments are never merged", path: "a.ts", code: "/* one */\n/* two */\n", want: []string{"R07 a.ts:1", "R07 a.ts:2"}},
		{name: "a line comment after a block comment is separate", path: "a.ts", code: "/* one */\n// two\n", want: []string{"R07 a.ts:1", "R07 a.ts:2"}},
		{name: "a multi-line block comment is reported at its first line", path: "a.ts", code: "a();\n/*\n * x\n */\nb();\n// after\n", want: []string{"R07 a.ts:2", "R07 a.ts:6"}},
		{name: "crlf line endings still form a run", path: "a.ts", code: "// one\r\n// two\r\na();\r\n// three\r\n", want: []string{"R07 a.ts:1", "R07 a.ts:4"}},
		{name: "rust doc comment run is one finding", path: "a.rs", code: "/// One.\n///\n/// Two.\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
