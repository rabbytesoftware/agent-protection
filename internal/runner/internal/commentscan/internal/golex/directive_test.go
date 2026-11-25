package golex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/golex"
	"testing"
)

func TestModuleDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment line is reported", code: "module x\n\n// pinned because\ngo 1.24\n", want: []int{3}},
		{name: "indirect marker passes", code: "require (\n\tgolang.org/x/text v0.3.0 // indirect\n)\n", want: nil},
		{name: "indirect alone on a line is reported", code: "// indirect\n", want: []int{1}},
		{name: "indirect with prose is reported", code: "require a v1 // indirect; pinned\n", want: []int{1}},
		{name: "retract rationale is reported", code: "retract v1.0.0 // published by mistake\n", want: []int{1}},
		{name: "slashes inside a quoted path pass", code: "replace a => \"../x//y\"\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			golex.Gomod(),
			tc.code,
			tc.want,
		)
	}
}

func TestSourceDirectives(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "build directive passes", code: "//go:build linux\n\npackage p\n", want: nil},
		{name: "line directive passes", code: "//line a.go:1\n", want: nil},
		{name: "export directive passes", code: "//export F\n", want: nil},
		{name: "legacy build line passes", code: "// +build linux\n", want: nil},
		{name: "trailing directive is reported", code: "x := 1 //go:noinline\n", want: []int{1}},
		{name: "directive with carriage return passes", code: "//go:build linux\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			golex.Gosource(),
			tc.code,
			tc.want,
		)
	}
}
