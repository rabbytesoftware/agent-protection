package commentpolicy_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/commentpolicy"
)

func TestCommentpolicyDirectives(
	t *testing.T,
) {
	cases := []struct {
		name    string
		comment string
		want    []string
	}{
		{name: "build constraint passes", comment: "//go:build linux", want: nil},
		{name: "generate passes", comment: "//go:generate stringer -type=T", want: nil},
		{name: "invented go directive is reported", comment: "//go:note the rate table is wrong", want: []string{"R07 p.go:3"}},
		{name: "bare go prefix is reported", comment: "//go:", want: []string{"R07 p.go:3"}},
		{name: "line directive with position passes", comment: "//line p.go:10", want: nil},
		{name: "line directive with column passes", comment: "//line p.go:10:2", want: nil},
		{name: "line prefix carrying prose is reported", comment: "//line HACK keep this", want: []string{"R07 p.go:3"}},
		{name: "export of a symbol passes", comment: "//export F", want: nil},
		{name: "export carrying prose is reported", comment: "//export F because cgo", want: []string{"R07 p.go:3"}},
		{name: "extern with a punctuated name is reported", comment: "//extern f.x", want: []string{"R07 p.go:3"}},
		{name: "legacy build line passes", comment: "// +build linux", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			commentpolicy.Commentpolicy([]source.File{parsed(
				t,
				"p.go",
				"package p\n\n"+tc.comment+"\nfunc F() {\n}\n",
			)}),
			tc.want,
		)
	}
}
