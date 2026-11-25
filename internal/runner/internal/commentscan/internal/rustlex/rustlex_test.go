package rustlex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/rustlex"
)

func TestRustlex(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "outer doc comment is reported", code: "/// Adds.\npub fn add() {}\n", want: []int{1}},
		{name: "inner doc comment is reported", code: "//! Crate docs.\nfn a() {}\n", want: []int{1}},
		{name: "block doc comment is reported", code: "fn a() {}\n/** Docs. */\nfn b() {}\n", want: []int{2}},
		{name: "safety comment is reported", code: "fn a() {\n    // SAFETY: checked above\n    unsafe { b() }\n}\n", want: []int{2}},
		{name: "license header is reported", code: "// Copyright the authors.\n// SPDX-License-Identifier: MIT\n\nfn a() {}\n", want: []int{1}},
		{name: "nested block comment is one comment", code: "/* outer /* inner */ still outer */\nfn a() {}\n", want: []int{1}},
		{name: "code after a nested block comment is lexed", code: "/* a /* b */ c */ let s = \"// no\";\n// real\n", want: []int{1, 2}},
		{name: "url in a string passes", code: "let u = \"https://x.dev\";\n", want: nil},
		{name: "multi-line string passes", code: "let u = \"line\n// not\n\";\n// real\n", want: []int{4}},
		{name: "raw string passes", code: "let u = r\"C:\\path // x\";\n", want: nil},
		{name: "raw string with hashes holds a quote", code: "let u = r#\"say \"// hi\"\"#; // real\n", want: []int{1}},
		{name: "raw string with two hashes holds a single hash closer", code: "let u = r##\"a \"# // b\"##;\n", want: nil},
		{name: "raw byte string passes", code: "let u = br#\"// x\"#;\n", want: nil},
		{name: "raw c string passes", code: "let u = cr\"// x\";\n", want: nil},
		{name: "byte string passes", code: "let u = b\"// x\";\n", want: nil},
		{name: "raw identifier is not a raw string", code: "let r#type = 1; // real\n", want: []int{1}},
		{name: "identifier ending in r is not a raw string", code: "let s = format!(\"{}\", bar); // real\n", want: []int{1}},
		{name: "char literal holding a quote", code: "let c = '\"'; let s = \"// x\";\n", want: nil},
		{name: "char literal holding a slash", code: "let c = '/'; let d = '/';\n", want: nil},
		{name: "escaped newline char", code: "let c = '\\n'; // real\n", want: []int{1}},
		{name: "escaped quote char", code: "let c = '\\''; let s = \"// x\";\n", want: nil},
		{name: "unicode escape char", code: "let c = '\\u{1F600}'; // real\n", want: []int{1}},
		{name: "multibyte char literal", code: "let c = 'é'; let s = \"// x\";\n", want: nil},
		{name: "byte char literal", code: "let c = b'\"'; let s = \"// x\";\n", want: nil},
		{name: "lifetime in a reference", code: "fn a<'a>(s: &'a str) -> &'a str { \"// x\" }\n", want: nil},
		{name: "static lifetime then a string", code: "const S: &'static str = \"don't // x\";\n", want: nil},
		{name: "loop label then a comment", code: "'outer: loop { break 'outer; } // real\n", want: []int{1}},
		{name: "shebang passes", code: "#!/usr/bin/env -S cargo +nightly -Zscript\nfn main() {}\n", want: nil},
		{name: "inner attribute on line one is code", code: "#![allow(dead_code)] // real\nfn a() {}\n", want: []int{1}},
		{name: "division is not a comment", code: "let a = b / c / d;\n", want: nil},
		{name: "unterminated raw string runs to the end", code: "let u = r#\"open\n// inside\n", want: nil},
		{name: "unterminated nested block comment is reported", code: "/* a /* b */\nfn a() {}\n", want: []int{1}},
		{name: "unterminated escaped char is a lone quote", code: "let c = '\\\n// real\n", want: []int{2}},
		{name: "quote at end of file is safe", code: "let c = '", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			rustlex.Rustlex(),
			tc.code,
			tc.want,
		)
	}
}

func TestRustlexRuns(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "rust doc comment run is one finding", code: "/// One.\n///\n/// Two.\nfn a() {}\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			rustlex.Rustlex(),
			tc.code,
			tc.want,
		)
	}
}

func TestRustlexMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		rustlex.Rustlex(),
		[]string{"a.rs", "src/lib.rs"},
		[]string{"a.rst", "rs"},
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
