package shlex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
	"testing"
)

func TestYaml(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment line is reported", code: "a: 1\n# note\n", want: []int{2}},
		{name: "trailing comment is reported", code: "a: 1 # why\n", want: []int{1}},
		{name: "hash in a url passes", code: "u: https://x.dev/#frag\n", want: nil},
		{name: "hash in quotes passes", code: "c: \"#fff\"\nd: '#000'\n", want: nil},
		{name: "apostrophe in a plain scalar", code: "t: don't panic\n# real\n", want: []int{2}},
		{name: "data block scalar is not lexed", code: "description: |\n  ## Heading\n  # not a comment\nnext: 1\n# real\n", want: []int{5}},
		{name: "folded block is data", code: "f: >-\n  # text\n", want: nil},
		{name: "list item block is data", code: "- |\n  # text\n- x\n", want: nil},
		{name: "run block is lexed as shell", code: "steps:\n  - run: |\n      echo \"#\" # why\n", want: []int{3}},
		{name: "heredoc inside a run block", code: "run: |\n  cat <<EOF\n  # body\n  EOF\n  # real\n", want: []int{5}},
		{name: "block ends at a dedent", code: "a:\n  b: |\n    # data\n  c: 1 # real\n", want: []int{4}},
		{name: "comment on a block header is reported", code: "d: | # why\n  text\n", want: []int{1}},
		{name: "blank lines stay inside a data block", code: "d: |\n  a\n\n  # b\nn: 1\n", want: nil},
		{name: "quoted run key is lexed as shell", code: "\"run\": |\n  x # y\n", want: []int{2}},
		{name: "block header at end of file", code: "d: |", want: nil},
		{name: "line that only holds an indicator is not a block", code: "|\n# real\n", want: []int{2}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			shlex.Yaml(),
			tc.code,
			tc.want,
		)
	}
}
