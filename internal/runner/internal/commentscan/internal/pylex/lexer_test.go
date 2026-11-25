package pylex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/pylex"
	"testing"
)

func TestHashLexer(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "trailing comment is reported", code: "x = 1  # why\n", want: []int{1}},
		{name: "hash in strings passes", code: "s = '#'\nt = \"#\"\n", want: nil},
		{name: "hash glued to code is a comment", code: "x=1#why\n", want: []int{1}},
		{name: "triple quoted string spans lines", code: "s = \"\"\"\n# not\n\"\"\"\n# real\n", want: []int{4}},
		{name: "single triple quotes span lines", code: "s = '''\n# not\n'''\n", want: nil},
		{name: "escaped quote inside a string", code: "s = 'it\\'s # not'\n", want: nil},
		{name: "escaped quote inside a triple string", code: "s = \"\"\"a \\\"\"\" b\"\"\" # real\n", want: []int{1}},
		{name: "unterminated triple string runs to the end", code: "s = \"\"\"open\n# inside\n", want: nil},
		{name: "unterminated string stops at the line end", code: "s = 'open\n# real\n", want: []int{2}},
		{name: "docstring is a string", code: "def f():\n    \"\"\"Docs # x.\"\"\"\n", want: nil},
		{name: "shebang passes", code: "#!/usr/bin/env python3\nprint(1)\n", want: nil},
		{name: "line continuation backslash", code: "x = 1 + \\\n  2 # real\n", want: []int{2}},
		{name: "begin end block is reported", code: "x = 1\n=begin\nprose # x\n=end\ny = 2\n", want: []int{2}},
		{name: "unterminated begin block is reported", code: "=begin\nprose\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			pylex.Python(),
			tc.code,
			tc.want,
		)
	}
}
