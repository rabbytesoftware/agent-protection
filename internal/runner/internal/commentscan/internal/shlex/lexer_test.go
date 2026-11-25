package shlex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
	"testing"
)

func TestShellLexer(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment line is reported", code: "#!/bin/sh\n# note\necho hi\n", want: []int{2}},
		{name: "shebang passes", code: "#!/usr/bin/env bash\necho hi\n", want: nil},
		{name: "trailing comment is reported", code: "echo hi # why\n", want: []int{1}},
		{name: "hash inside double quotes passes", code: "echo \"a # b\"\n", want: nil},
		{name: "hash inside single quotes passes", code: "echo 'a # b'\n", want: nil},
		{name: "hash glued to a word passes", code: "echo a#b\n", want: nil},
		{name: "parameter length passes", code: "echo ${#items[@]} $#\n", want: nil},
		{name: "prefix strip passes", code: "V=\"${LATEST#v}\"; W=${X##*/}\n", want: nil},
		{name: "escaped hash passes", code: "echo \\# not\n", want: nil},
		{name: "apostrophe inside a word does not open a quote", code: "echo don't\n# real\n", want: []int{2}},
		{name: "assignment opens a single quote", code: "X='a # b'\n", want: nil},
		{name: "double quoted string spans lines", code: "echo \"one\n# not\n\"\n# real\n", want: []int{4}},
		{name: "single quoted script spans lines", code: "awk '\n# not\n'\n", want: nil},
		{name: "comment after a semicolon is reported", code: "a;# why\n", want: []int{1}},
		{name: "consecutive comment lines are one finding", code: "# one\n# two\necho\n", want: []int{1}},
		{name: "unterminated double quote runs to the end", code: "echo \"open\n# inside\n", want: nil},
		{name: "unterminated single quote runs to the end", code: "echo 'open\n# inside\n", want: nil},
		{name: "trailing backslash at end of file is safe", code: "echo \\", want: nil},
		{name: "comment at end of file without newline", code: "echo # end", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			shlex.Shell(),
			tc.code,
			tc.want,
		)
	}
}
