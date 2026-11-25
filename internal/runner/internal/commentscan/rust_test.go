package commentscan_test

import "testing"

func TestRust(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "outer doc comment is reported", path: "a.rs", code: "/// Adds.\npub fn add() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "inner doc comment is reported", path: "a.rs", code: "//! Crate docs.\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "block doc comment is reported", path: "a.rs", code: "fn a() {}\n/** Docs. */\nfn b() {}\n", want: []string{"R07 a.rs:2"}},
		{name: "safety comment is reported", path: "a.rs", code: "fn a() {\n    // SAFETY: checked above\n    unsafe { b() }\n}\n", want: []string{"R07 a.rs:2"}},
		{name: "license header is reported", path: "a.rs", code: "// Copyright the authors.\n// SPDX-License-Identifier: MIT\n\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "nested block comment is one comment", path: "a.rs", code: "/* outer /* inner */ still outer */\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "code after a nested block comment is lexed", path: "a.rs", code: "/* a /* b */ c */ let s = \"// no\";\n// real\n", want: []string{"R07 a.rs:1", "R07 a.rs:2"}},
		{name: "url in a string passes", path: "a.rs", code: "let u = \"https://x.dev\";\n", want: nil},
		{name: "multi-line string passes", path: "a.rs", code: "let u = \"line\n// not\n\";\n// real\n", want: []string{"R07 a.rs:4"}},
		{name: "raw string passes", path: "a.rs", code: "let u = r\"C:\\path // x\";\n", want: nil},
		{name: "raw string with hashes holds a quote", path: "a.rs", code: "let u = r#\"say \"// hi\"\"#; // real\n", want: []string{"R07 a.rs:1"}},
		{name: "raw string with two hashes holds a single hash closer", path: "a.rs", code: "let u = r##\"a \"# // b\"##;\n", want: nil},
		{name: "raw byte string passes", path: "a.rs", code: "let u = br#\"// x\"#;\n", want: nil},
		{name: "raw c string passes", path: "a.rs", code: "let u = cr\"// x\";\n", want: nil},
		{name: "byte string passes", path: "a.rs", code: "let u = b\"// x\";\n", want: nil},
		{name: "raw identifier is not a raw string", path: "a.rs", code: "let r#type = 1; // real\n", want: []string{"R07 a.rs:1"}},
		{name: "identifier ending in r is not a raw string", path: "a.rs", code: "let s = format!(\"{}\", bar); // real\n", want: []string{"R07 a.rs:1"}},
		{name: "char literal holding a quote", path: "a.rs", code: "let c = '\"'; let s = \"// x\";\n", want: nil},
		{name: "char literal holding a slash", path: "a.rs", code: "let c = '/'; let d = '/';\n", want: nil},
		{name: "escaped newline char", path: "a.rs", code: "let c = '\\n'; // real\n", want: []string{"R07 a.rs:1"}},
		{name: "escaped quote char", path: "a.rs", code: "let c = '\\''; let s = \"// x\";\n", want: nil},
		{name: "unicode escape char", path: "a.rs", code: "let c = '\\u{1F600}'; // real\n", want: []string{"R07 a.rs:1"}},
		{name: "multibyte char literal", path: "a.rs", code: "let c = 'é'; let s = \"// x\";\n", want: nil},
		{name: "byte char literal", path: "a.rs", code: "let c = b'\"'; let s = \"// x\";\n", want: nil},
		{name: "lifetime in a reference", path: "a.rs", code: "fn a<'a>(s: &'a str) -> &'a str { \"// x\" }\n", want: nil},
		{name: "static lifetime then a string", path: "a.rs", code: "const S: &'static str = \"don't // x\";\n", want: nil},
		{name: "loop label then a comment", path: "a.rs", code: "'outer: loop { break 'outer; } // real\n", want: []string{"R07 a.rs:1"}},
		{name: "shebang passes", path: "a.rs", code: "#!/usr/bin/env -S cargo +nightly -Zscript\nfn main() {}\n", want: nil},
		{name: "inner attribute on line one is code", path: "a.rs", code: "#![allow(dead_code)] // real\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "division is not a comment", path: "a.rs", code: "let a = b / c / d;\n", want: nil},
		{name: "unterminated raw string runs to the end", path: "a.rs", code: "let u = r#\"open\n// inside\n", want: nil},
		{name: "unterminated nested block comment is reported", path: "a.rs", code: "/* a /* b */\nfn a() {}\n", want: []string{"R07 a.rs:1"}},
		{name: "unterminated escaped char is a lone quote", path: "a.rs", code: "let c = '\\\n// real\n", want: []string{"R07 a.rs:2"}},
		{name: "quote at end of file is safe", path: "a.rs", code: "let c = '", want: nil},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
