package comments_test

import "testing"

func TestScript(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "jsdoc block is reported", path: "a.ts", code: "/**\n * Adds.\n */\nexport function add() {}\n", want: []string{"R07 a.ts:1"}},
		{name: "ts-ignore is reported", path: "a.ts", code: "// @ts-ignore\nconst a: number = 'x';\n", want: []string{"R07 a.ts:1"}},
		{name: "eslint-disable is reported", path: "a.ts", code: "/* eslint-disable no-console */\nconsole.log(1);\n", want: []string{"R07 a.ts:1"}},
		{name: "eslint-disable-next-line is reported", path: "a.js", code: "a();\n// eslint-disable-next-line no-console\nconsole.log(1);\n", want: []string{"R07 a.js:2"}},
		{name: "license header is reported", path: "a.ts", code: "/*!\n * MIT License\n */\nexport {};\n", want: []string{"R07 a.ts:1"}},
		{name: "jsx comment is reported", path: "a.tsx", code: "const v = (\n  <div>\n    {/* hidden */}\n  </div>\n);\n", want: []string{"R07 a.tsx:3"}},
		{name: "triple slash reference passes", path: "a.ts", code: "/// <reference types=\"node\" />\n///<reference path=\"x.d.ts\" />\nexport {};\n", want: nil},
		{name: "triple slash prose is reported", path: "a.ts", code: "/// note\nexport {};\n", want: []string{"R07 a.ts:1"}},
		{name: "hash pure annotation passes", path: "a.js", code: "const a = /*#__PURE__*/ make();\n", want: nil},
		{name: "at pure annotation passes", path: "a.js", code: "const a = /*@__PURE__*/ make();\n", want: nil},
		{name: "other annotation comments are reported", path: "a.js", code: "import(/* webpackChunkName: \"x\" */ './x');\n", want: []string{"R07 a.js:1"}},
		{name: "template literal with a url passes", path: "a.ts", code: "const u = `https://${host}/a`;\n", want: nil},
		{name: "template literal spans lines", path: "a.ts", code: "const u = `\n// not a comment\n/* nor this */\n`;\n// real\n", want: []string{"R07 a.ts:5"}},
		{name: "template expression with a string holding a backtick and brace", path: "a.ts", code: "const u = `a ${f('`}', \"}\")} // still template`;\n", want: nil},
		{name: "template expression with an object literal", path: "a.ts", code: "const u = `a ${JSON.stringify({ b: { c: 1 } })} // still`;\n", want: nil},
		{name: "nested template literal", path: "a.ts", code: "const u = `a ${`b ${`c // deep`}`} // still`; // real\n", want: []string{"R07 a.ts:1"}},
		{name: "comment inside a template expression is reported", path: "a.ts", code: "const u = `a ${\n  x // why\n}`;\n", want: []string{"R07 a.ts:2"}},
		{name: "escaped backtick keeps the template open", path: "a.ts", code: "const u = `a \\` // still`;\n", want: nil},
		{name: "escaped dollar brace is template text", path: "a.ts", code: "const u = `a \\${ // still`;\n", want: nil},
		{name: "code after a closed template is lexed", path: "a.ts", code: "const u = `${a}`; // real\n", want: []string{"R07 a.ts:1"}},
		{name: "braces in code before a comment", path: "a.ts", code: "function f() {\n  if (a) { b(); }\n}\n// real\n", want: []string{"R07 a.ts:4"}},
		{name: "url in jsx text passes", path: "a.tsx", code: "const v = <a>see https://x.dev/a</a>;\n", want: nil},
		{name: "comment after a url in jsx text is reported", path: "a.tsx", code: "const v = <a>http://x</a>; // real\n", want: []string{"R07 a.tsx:1"}},
		{name: "comment after a spaced colon is reported", path: "a.ts", code: "const o = { a: // why\n  1 };\n", want: []string{"R07 a.ts:1"}},
		{name: "unterminated template runs to the end", path: "a.ts", code: "const u = `open\n// inside\n", want: nil},
		{name: "stray closing brace does not stop scanning", path: "a.ts", code: "}\n// real\n", want: []string{"R07 a.ts:2"}},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
