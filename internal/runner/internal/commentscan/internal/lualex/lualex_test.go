package lualex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lualex"
)

func TestLua(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "line comment is reported", code: "-- workaround: lua hack\nreturn 1\n", want: []int{1}},
		{name: "trailing comment is reported", code: "local a = 1 -- why\n", want: []int{1}},
		{name: "block comment is reported", code: "--[[ why\nstill ]]\nreturn 1\n", want: []int{1}},
		{name: "leveled block comment is reported", code: "--[==[ a ]] b ]==]\nreturn 1\n", want: []int{1}},
		{name: "dashes in strings pass", code: "local a = \"--x\" .. '--y'\n", want: nil},
		{name: "dashes in a long string pass", code: "local a = [[\n-- not\n]]\n-- real\n", want: []int{4}},
		{name: "leveled long string passes", code: "local a = [=[ ]] -- not ]=]\n", want: nil},
		{name: "index brackets are code", code: "local a = t[1] - b[2]\n", want: nil},
		{name: "shebang passes", code: "#!/usr/bin/env lua\nprint(1)\n", want: nil},
		{name: "unterminated block runs to the end", code: "--[[ open\nreturn 1\n", want: []int{1}},
		{name: "open bracket at the end is safe", code: "local a = t[", want: nil},
		{name: "unterminated level is not long", code: "local a = t[==\n-- real\n", want: []int{2}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			lualex.Lua(),
			tc.code,
			tc.want,
		)
	}
}

func TestLuaMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		lualex.Lua(),
		[]string{"a.lua", "B.LUA"},
		[]string{"a.luac", "lua"},
	)
}

func TestLuaInterpreters(
	t *testing.T,
) {
	if !slices.Contains(
		lualex.Lua().Interpreters,
		"lua",
	) {
		t.Error("Lua() does not interpret lua scripts")
	}
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
