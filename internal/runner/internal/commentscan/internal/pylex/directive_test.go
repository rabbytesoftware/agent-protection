package pylex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/pylex"
	"testing"
)

func TestPythonDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "coding declaration on line one passes", code: "# -*- coding: utf-8 -*-\nx = 1\n", want: nil},
		{name: "coding declaration on line two passes", code: "#!/usr/bin/env python3\n# coding=utf-8\n", want: nil},
		{name: "vim encoding declaration passes", code: "# vim: set fileencoding=utf-8 :\n", want: nil},
		{name: "coding declaration on line three is reported", code: "x = 1\ny = 2\n# coding: utf-8\n", want: []int{3}},
		{name: "coding declaration with prose is reported", code: "# HACK coding: utf-8\n", want: []int{1}},
		{name: "type ignore suppression is reported", code: "x = f()  # type: ignore\n", want: []int{1}},
		{name: "noqa suppression is reported", code: "import os  # noqa: F401\n", want: []int{1}},
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

func TestRubyDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "frozen string literal passes", code: "# frozen_string_literal: true\nputs 1\n", want: nil},
		{name: "magic comments after a shebang pass", code: "#!/usr/bin/env ruby\n# encoding: utf-8\n# frozen_string_literal: true\n", want: nil},
		{name: "magic comment after code is reported", code: "puts 1\n# frozen_string_literal: true\n", want: []int{2}},
		{name: "prose after a magic comment is reported", code: "# frozen_string_literal: true\n\n# Greets.\ndef hi; end\n", want: []int{3}},
		{name: "interpolation inside a string passes", code: "puts \"#{name} # not\"\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			pylex.Ruby(),
			tc.code,
			tc.want,
		)
	}
}

func TestTomlDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "schema directive on line one passes", code: "#:schema ./schema.json\na = 1\n", want: nil},
		{name: "schema directive later is reported", code: "a = 1\n#:schema x.json\n", want: []int{2}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			pylex.Toml(),
			tc.code,
			tc.want,
		)
	}
}
