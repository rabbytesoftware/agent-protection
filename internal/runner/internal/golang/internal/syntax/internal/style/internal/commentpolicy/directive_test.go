package commentpolicy_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/commentpolicy"
)

func TestCommentpolicyDirectives(
	t *testing.T,
) {
	cases := []struct {
		name    string
		comment string
		want    []string
	}{
		{name: "generate passes", comment: "//go:generate stringer -type=T", want: nil},
		{name: "generate echo carrying prose is reported", comment: "//go:generate echo factor stays 1 because billing double counts", want: []string{"R07 p.go:3"}},
		{name: "generate true carrying prose is reported", comment: "//go:generate true this is a hack", want: []string{"R07 p.go:3"}},
		{name: "generate colon carrying prose is reported", comment: "//go:generate : hack", want: []string{"R07 p.go:3"}},
		{name: "generate printf carrying prose is reported", comment: "//go:generate printf hack", want: []string{"R07 p.go:3"}},
		{name: "generate false carrying prose is reported", comment: "//go:generate false hack", want: []string{"R07 p.go:3"}},
		{name: "generate echo by path is reported", comment: "//go:generate /bin/echo hack", want: []string{"R07 p.go:3"}},
		{name: "generate alias to echo is reported", comment: "//go:generate -command note echo", want: []string{"R07 p.go:3"}},
		{name: "generate alias to a tool passes", comment: "//go:generate -command yacc go tool yacc", want: nil},
		{name: "incomplete generate alias is reported", comment: "//go:generate -command x", want: []string{"R07 p.go:3"}},
		{name: "bare generate is reported", comment: "//go:generate", want: []string{"R07 p.go:3"}},
		{name: "noinline passes", comment: "//go:noinline", want: nil},
		{name: "noinline carrying prose is reported", comment: "//go:noinline TODO: inlining breaks the profiler", want: []string{"R07 p.go:3"}},
		{name: "nosplit carrying prose is reported", comment: "//go:nosplit because the stack is small", want: []string{"R07 p.go:3"}},
		{name: "linkname passes", comment: "//go:linkname f runtime.f", want: nil},
		{name: "linkname carrying prose is reported", comment: "//go:linkname f runtime.f because we must", want: []string{"R07 p.go:3"}},
		{name: "bare linkname is reported", comment: "//go:linkname", want: []string{"R07 p.go:3"}},
		{name: "wasmimport passes", comment: "//go:wasmimport env f", want: nil},
		{name: "wasmimport carrying prose is reported", comment: "//go:wasmimport env f extra words", want: []string{"R07 p.go:3"}},
		{name: "wasmexport passes", comment: "//go:wasmexport f", want: nil},
		{name: "wasmexport carrying prose is reported", comment: "//go:wasmexport f extra", want: []string{"R07 p.go:3"}},
		{name: "debug setting passes", comment: "//go:debug panicnil=1", want: nil},
		{name: "debug prose is reported", comment: "//go:debug keep this", want: []string{"R07 p.go:3"}},
		{name: "embed without a pattern is reported", comment: "//go:embed", want: []string{"R07 p.go:3"}},
		{name: "cgo ldflag passes", comment: "//go:cgo_ldflag -lm", want: nil},
		{name: "invented go directive is reported", comment: "//go:note the rate table is wrong", want: []string{"R07 p.go:3"}},
		{name: "bare go prefix is reported", comment: "//go:", want: []string{"R07 p.go:3"}},
		{name: "build after the package clause is reported", comment: "//go:build linux", want: []string{"R07 p.go:3"}},
		{name: "legacy build after the package clause is reported", comment: "// +build linux", want: []string{"R07 p.go:3"}},
		{name: "line directive with position passes", comment: "//line p.go:10", want: nil},
		{name: "line directive with column passes", comment: "//line p.go:10:2", want: nil},
		{name: "line prefix carrying prose is reported", comment: "//line HACK keep this", want: []string{"R07 p.go:3"}},
		{name: "export of a symbol passes", comment: "//export F", want: nil},
		{name: "export carrying prose is reported", comment: "//export F because cgo", want: []string{"R07 p.go:3"}},
		{name: "extern with a punctuated name is reported", comment: "//extern f.x", want: []string{"R07 p.go:3"}},
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

func TestCommentpolicyPlacement(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "build constraint before the package passes", path: "p.go", code: "//go:build linux\n\npackage p\n", want: nil},
		{name: "build constraint carrying prose is reported", path: "p.go", code: "//go:build linux TODO: fix\n\npackage p\n", want: []string{"R07 p.go:1"}},
		{name: "legacy build paired with go build passes", path: "p.go", code: "//go:build linux\n// +build linux\n\npackage p\n", want: nil},
		{name: "legacy build without go build is reported", path: "p.go", code: "// +build linux because\n\npackage p\n", want: []string{"R07 p.go:1"}},
		{name: "generate inside a function body is reported", path: "p.go", code: "package p\n\nfunc F() {\n\t//go:generate stringer -type=T\n}\n", want: []string{"R07 p.go:4"}},
		{name: "generate inside a struct is reported", path: "p.go", code: "package p\n\ntype T struct {\n\t//go:generate stringer\n\tA int\n}\n", want: []string{"R07 p.go:4"}},
		{name: "generate between declarations passes", path: "p.go", code: "package p\n\nvar A = 1\n\n//go:generate stringer -type=T\n\nvar B = 2\n", want: nil},
		{name: "red team square file is reported", path: "p.go", code: "package square\n\ntype square struct {\n\tfactor int\n}\n\n//go:generate echo factor stays 1 because the legacy billing API double counts otherwise\nfunc Square() (squarer *square) {\n\treturn &square{factor: 1}\n}\n\n//go:noinline TODO: inlining breaks the profiler, revisit after the Q3 migration\nfunc (s *square) Square(\n\tvalue int,\n) (result int) {\n\t//go:generate true this branch is a hack until finance signs off on the new rounding\n\treturn value * value * s.factor\n}\n", want: []string{"R07 p.go:7", "R07 p.go:12", "R07 p.go:16"}},
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
