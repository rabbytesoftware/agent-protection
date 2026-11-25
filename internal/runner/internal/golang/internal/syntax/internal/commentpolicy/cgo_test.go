package commentpolicy_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/commentpolicy"
)

func TestCommentpolicyPreamble(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "include passes", path: "p.go", code: "package p\n\n// #include <stdio.h>\nimport \"C\"\n", want: nil},
		{name: "c line comment in a line preamble is reported", path: "p.go", code: "package p\n\n// // The factor is pinned.\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "each commented preamble line is reported", path: "p.go", code: "package square\n\n// // The factor is pinned to 1 because the legacy billing API double counts.\n// // TODO: drop this once finance signs off; nobody remembers why it exists.\nimport \"C\"\n\nconst Factor = 1\n", want: []string{"R07 p.go:3", "R07 p.go:4"}},
		{name: "c block comment is reported", path: "p.go", code: "package p\n\n// /* why */\n// #include <x.h>\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "c comment in a block preamble is reported", path: "p.go", code: "package p\n\n/*\n#include <x.h>\n// why\n*/\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "slashes inside a c string pass", path: "p.go", code: "package p\n\n// #define URL \"http://x\"\nimport \"C\"\n", want: nil},
		{name: "slashes inside char literals pass", path: "p.go", code: "package p\n\n// char c = '/'; char d = '/';\nimport \"C\"\n", want: nil},
		{name: "escaped quote inside a c string passes", path: "p.go", code: "package p\n\n// #define Q \"a\\\"//b\"\nimport \"C\"\n", want: nil},
		{name: "unterminated c string passes", path: "p.go", code: "package p\n\n// #define Q \"abc\nimport \"C\"\n", want: nil},
		{name: "trailing c comment after code is reported", path: "p.go", code: "package p\n\n// int x; // why\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "disabled block is reported", path: "p.go", code: "package p\n\n// #if 0\n// prose here\n// #endif\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "disabled block in a block preamble is reported", path: "p.go", code: "package p\n\n/*\n#if 0\nprose\n#endif\n*/\nimport \"C\"\n", want: []string{"R07 p.go:3"}},
		{name: "if with a real condition passes", path: "p.go", code: "package p\n\n// #if 01 || defined(X)\n// #endif\nimport \"C\"\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			commentpolicy.Commentpolicy([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}
