package shlex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
	"testing"
)

func TestHeredoc(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "heredoc body is not lexed", code: "cat <<EOF\n# Title\nEOF\n# real\n", want: []int{4}},
		{name: "single quoted delimiter", code: "cat <<'EOF'\n# Title\nEOF\n", want: nil},
		{name: "double quoted delimiter", code: "cat <<\"END\"\n# x\nEND\n# real\n", want: []int{4}},
		{name: "dash form with an indented terminator", code: "cat <<-EOF\n\t# x\n\tEOF\n# real\n", want: []int{4}},
		{name: "escaped delimiter", code: "cat <<\\EOF\n# x\nEOF\n", want: nil},
		{name: "space before the delimiter", code: "cat << EOF\n# x\nEOF\n", want: nil},
		{name: "comment on the heredoc line is reported", code: "cat <<EOF # why\nbody\nEOF\n", want: []int{1}},
		{name: "two heredocs on one line", code: "paste <<A <<B\n# a\nA\n# b\nB\n# real\n", want: []int{6}},
		{name: "here-string is not a heredoc", code: "cat <<< \"x\" # why\n", want: []int{1}},
		{name: "shift operator is not a heredoc", code: "echo $((1<<2))\n# real\n", want: []int{2}},
		{name: "unterminated heredoc runs to the end", code: "cat <<EOF\n# x\n", want: nil},
		{name: "unterminated quoted delimiter is not a heredoc", code: "cat <<'EOF\n# inside\n", want: nil},
		{name: "heredoc operator at end of file is safe", code: "cat <<", want: nil},
		{name: "heredoc on the last line without newline", code: "cat <<EOF", want: nil},
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
