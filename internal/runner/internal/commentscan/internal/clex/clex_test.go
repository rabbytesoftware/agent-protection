package clex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/clex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func TestProto(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "line comment is reported", code: "syntax = \"proto3\";\n// skip validation\npackage a;\n", want: []int{2}},
		{name: "syntax declaration is code", code: "syntax = \"proto3\";\npackage a;\n", want: nil},
		{name: "buf lint suppression is reported", code: "// buf:lint:ignore FIELD_LOWER_SNAKE_CASE\nmessage A {}\n", want: []int{1}},
		{name: "block comment is reported", code: "message A { /* why */ int64 v = 1; }\n", want: []int{1}},
		{name: "slashes in a string pass", code: "option go_package = \"a//b\";\n", want: nil},
		{name: "single quoted slashes pass", code: "option a = 'x//y';\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			clex.Proto(),
			tc.code,
			tc.want,
		)
	}
}
func TestCss(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "block comment is reported", code: "/* workaround for safari */\np { color: red; }\n", want: []int{1}},
		{name: "double slash is not a css comment", code: "a { background: url(http://x/y.png); }\n", want: nil},
		{name: "comment marker in a string passes", code: "a::before { content: \"/* no */\"; }\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			clex.Css(),
			tc.code,
			tc.want,
		)
	}
}
func TestScss(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "line comment is reported", code: "// hack\n$a: 1;\n", want: []int{1}},
		{name: "unquoted url passes", code: "a { background: url(https://x/y.png); }\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			clex.Scss(),
			tc.code,
			tc.want,
		)
	}
}
func TestJsonc(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "line comment is reported", code: "{\n  // keep this off\n  \"a\": 1\n}\n", want: []int{2}},
		{name: "block comment is reported", code: "{ /* why */ \"a\": 1 }\n", want: []int{1}},
		{name: "urls in strings pass", code: "{\"u\": \"https://x//y\"}\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			clex.Jsonc(),
			tc.code,
			tc.want,
		)
	}
}
func TestHcl(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "hash comment is reported", code: "# disabled auth\nresource \"a\" \"b\" {}\n", want: []int{1}},
		{name: "slash comment is reported", code: "a = 1 // why\n", want: []int{1}},
		{name: "hash in a string passes", code: "a = \"#not\"\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			clex.Hcl(),
			tc.code,
			tc.want,
		)
	}
}

func TestClexMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		clex.Proto(),
		[]string{"a.proto"},
		[]string{"a.protobuf", "proto"},
	)
}

func TestClexStyleMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		clex.Scss(),
		[]string{"a.scss", "b.less", "C.SASS"},
		[]string{"a.css"},
	)
}

func TestClexDataMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		clex.Jsonc(),
		[]string{"a.json", "tsconfig.jsonc", "a.json5"},
		[]string{"a.yaml"},
	)
}

func TestClexInfraMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		clex.Hcl(),
		[]string{"main.tf", "a.tfvars", "a.hcl"},
		[]string{"a.tfstate"},
	)
}

func TestClexSheetMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		clex.Css(),
		[]string{"a.css"},
		[]string{"a.scss"},
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
