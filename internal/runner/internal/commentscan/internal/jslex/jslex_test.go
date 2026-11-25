package jslex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/jslex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestJslex(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "jsdoc block is reported", code: "/**\n * Adds.\n */\nexport function add() {}\n", want: []int{1}},
		{name: "ts-ignore is reported", code: "// @ts-ignore\nconst a: number = 'x';\n", want: []int{1}},
		{name: "eslint-disable is reported", code: "/* eslint-disable no-console */\nconsole.log(1);\n", want: []int{1}},
		{name: "eslint-disable-next-line is reported", code: "a();\n// eslint-disable-next-line no-console\nconsole.log(1);\n", want: []int{2}},
		{name: "license header is reported", code: "/*!\n * MIT License\n */\nexport {};\n", want: []int{1}},
		{name: "jsx comment is reported", code: "const v = (\n  <div>\n    {/* hidden */}\n  </div>\n);\n", want: []int{3}},
		{name: "triple slash reference passes", code: "/// <reference types=\"node\" />\n///<reference path=\"x.d.ts\" />\nexport {};\n", want: nil},
		{name: "triple slash prose is reported", code: "/// note\nexport {};\n", want: []int{1}},
		{name: "hash pure annotation passes", code: "const a = /*#__PURE__*/ make();\n", want: nil},
		{name: "at pure annotation passes", code: "const a = /*@__PURE__*/ make();\n", want: nil},
		{name: "other annotation comments are reported", code: "import(/* webpackChunkName: \"x\" */ './x');\n", want: []int{1}},
		{name: "template literal with a url passes", code: "const u = `https://${host}/a`;\n", want: nil},
		{name: "template literal spans lines", code: "const u = `\n// not a comment\n/* nor this */\n`;\n// real\n", want: []int{5}},
		{name: "template expression with a string holding a backtick and brace", code: "const u = `a ${f('`}', \"}\")} // still template`;\n", want: nil},
		{name: "template expression with an object literal", code: "const u = `a ${JSON.stringify({ b: { c: 1 } })} // still`;\n", want: nil},
		{name: "nested template literal", code: "const u = `a ${`b ${`c // deep`}`} // still`; // real\n", want: []int{1}},
		{name: "comment inside a template expression is reported", code: "const u = `a ${\n  x // why\n}`;\n", want: []int{2}},
		{name: "escaped backtick keeps the template open", code: "const u = `a \\` // still`;\n", want: nil},
		{name: "escaped dollar brace is template text", code: "const u = `a \\${ // still`;\n", want: nil},
		{name: "code after a closed template is lexed", code: "const u = `${a}`; // real\n", want: []int{1}},
		{name: "braces in code before a comment", code: "function f() {\n  if (a) { b(); }\n}\n// real\n", want: []int{4}},
		{name: "a label glued to a comment is reported", code: "const o = { note:// hidden\n};\n", want: []int{1}},
		{name: "url in jsx text passes", code: "const v = <a>see https://x.dev/a</a>;\n", want: nil},
		{name: "comment after a url in jsx text is reported", code: "const v = <a>http://x</a>; // real\n", want: []int{1}},
		{name: "comment after a spaced colon is reported", code: "const o = { a: // why\n  1 };\n", want: []int{1}},
		{name: "unterminated template runs to the end", code: "const u = `open\n// inside\n", want: nil},
		{name: "stray closing brace does not stop scanning", code: "}\n// real\n", want: []int{2}},
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

func TestJslexLiterals(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "url in a double quoted string passes", code: "const u = \"https://x.dev\";\n", want: nil},
		{name: "comment opener in a single quoted string passes", code: "const u = '/* not */ // not';\n", want: nil},
		{name: "escaped quote keeps the string open", code: "const u = \"a\\\" // still string\";\n", want: nil},
		{name: "escaped backslash closes the string", code: "const u = \"a\\\\\"; // real\n", want: []int{1}},
		{name: "unterminated ts string stops at the line end", code: "const u = 'oops\n// real\n", want: []int{2}},
		{name: "shebang on line one passes", code: "#!/usr/bin/env node\nrun();\n", want: nil},
		{name: "shebang with a double slash passes", code: "#!/usr/bin/env -S node --x=//y\nrun();\n", want: nil},
		{name: "unterminated block comment is reported", code: "a();\n/* open\n", want: []int{2}},
		{name: "trailing backslash at end of file is safe", code: "const u = '\\", want: nil},
		{name: "comment at end of file without newline is reported", code: "a(); // end", want: []int{1}},
		{name: "unicode identifiers do not confuse the lexer", code: "const café = 'x'; // c\n", want: []int{1}},
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

func TestJslexRuns(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "consecutive line comments are one finding", code: "// one\n// two\n  // three\nconst a = 1;\n", want: []int{1}},
		{name: "a blank line splits a run", code: "// one\n\n// two\n", want: []int{1, 3}},
		{name: "code between line comments splits a run", code: "// one\nconst a = 1;\n// two\n", want: []int{1, 3}},
		{name: "trailing comments on code lines are separate", code: "a(); // one\nb(); // two\n", want: []int{1, 2}},
		{name: "block comments are never merged", code: "/* one */\n/* two */\n", want: []int{1, 2}},
		{name: "a line comment after a block comment is separate", code: "/* one */\n// two\n", want: []int{1, 2}},
		{name: "a multi-line block comment is reported at its first line", code: "a();\n/*\n * x\n */\nb();\n// after\n", want: []int{2, 6}},
		{name: "crlf line endings still form a run", code: "// one\r\n// two\r\na();\r\n// three\r\n", want: []int{1, 4}},
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

func TestJslexMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		jslex.Jslex(),
		[]string{"a.ts", "b.tsx", "c.js", "d.jsx", "e.mjs", "f.cjs", "g.mts", "h.cts", "types/a.d.ts"},
		[]string{"a.py", "a.json", "ts"},
	)
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
