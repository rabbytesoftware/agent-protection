package shlex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
	"testing"
)

func TestShellDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "shellcheck directive passes", code: "# shellcheck disable=SC2034\nX=1\n", want: nil},
		{name: "shellcheck directive with several settings passes", code: "# shellcheck source=lib.sh shell=bash\n", want: nil},
		{name: "shellcheck directive carrying prose is reported", code: "# shellcheck disable=SC2034 because unused\n", want: []int{1}},
		{name: "trailing shellcheck directive is reported", code: "X=1 # shellcheck disable=SC2034\n", want: []int{1}},
		{name: "bare shellcheck word is reported", code: "# shellcheck\n", want: []int{1}},
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

func TestYamlDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "schema directive passes", code: "# yaml-language-server: $schema=https://x.dev/s.json\na: 1\n", want: nil},
		{name: "schema directive with prose is reported", code: "# yaml-language-server: $schema=x because\n", want: []int{1}},
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

func TestDockerfileDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "syntax directive passes", code: "# syntax=docker/dockerfile:1\nFROM alpine\n", want: nil},
		{name: "escape and check directives pass", code: "# escape=`\n# check=skip=all\nFROM alpine\n", want: nil},
		{name: "directive after an instruction is reported", code: "FROM alpine\n# syntax=docker/dockerfile:1\n", want: []int{2}},
		{name: "prose comment is reported", code: "# base image\nFROM alpine\n", want: []int{1}},
		{name: "prose after directives is reported", code: "# syntax=x\n# why\nFROM a\n", want: []int{2}},
		{name: "trailing shell comment in a run instruction", code: "RUN make # why\n", want: []int{1}},
		{name: "heredoc in a run instruction", code: "RUN <<EOF\n# body\nEOF\n", want: nil},
		{name: "directives up to the end of file", code: "# syntax=x", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			shlex.Dockerfile(),
			tc.code,
			tc.want,
		)
	}
}
