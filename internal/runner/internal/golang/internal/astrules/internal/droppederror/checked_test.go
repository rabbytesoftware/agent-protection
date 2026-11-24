package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDroppederrorCheckedErrors(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "fallback value returned in place of the error is reported", code: "package p\n\nfunc f() bool {\n\t_, err := g()\n\tif err != nil {\n\t\treturn false\n\t}\n\n\treturn true\n}\n", want: []string{"R26 p.go:5"}},
		{name: "sentinel returned in place of the error is reported", code: "package p\n\nfunc f() error {\n\tif err := g(); err != nil {\n\t\treturn ErrInvalid\n\t}\n\n\treturn nil\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error joined with another condition is reported", code: "package p\n\nfunc f() bool {\n\tn, err := g()\n\tif err != nil ||\n\t\tn == 0 {\n\t\treturn false\n\t}\n\n\treturn true\n}\n", want: []string{"R26 p.go:5"}},
		{name: "error skipped with continue is reported", code: "package p\n\nfunc f() {\n\tfor range xs {\n\t\tif err := g(); err != nil {\n\t\t\tcontinue\n\t\t}\n\t}\n}\n", want: []string{"R26 p.go:5"}},
		{name: "error dropped inside a callback parameter is reported", code: "package p\n\nfunc f() {\n\twalk(func(walkErr error) error {\n\t\tif walkErr != nil {\n\t\t\treturn nil\n\t\t}\n\n\t\treturn nil\n\t})\n}\n", want: []string{"R26 p.go:5"}},
		{name: "nil on the left of the comparison is reported", code: "package p\n\nfunc f() {\n\tif err := g(); nil != err {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error logged before leaving passes", code: "package p\n\nfunc f() {\n\tif err := g(); err != nil {\n\t\tlog.Print(err)\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "error classified with errors.Is in the condition passes", code: "package p\n\nfunc f() error {\n\tif err := g(); err != nil && errors.Is(err, ErrMissing) {\n\t\treturn nil\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "error recorded before the check passes", code: "package p\n\nfunc f() {\n\terr := g()\n\tlast = err\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "block that falls through passes", code: "package p\n\nfunc f() error {\n\terr := g()\n\tif err != nil {\n\t\tcleanup()\n\t}\n\n\treturn err\n}\n", want: nil},
		{name: "retry loop returning the last error passes", code: "package p\n\nfunc f() (err error) {\n\tfor range 3 {\n\t\terr = g()\n\t\tif err != nil {\n\t\t\tcontinue\n\t\t}\n\n\t\treturn nil\n\t}\n\n\treturn err\n}\n", want: nil},
		{name: "break followed by a use after the loop passes", code: "package p\n\nfunc f() (err error) {\n\tfor {\n\t\terr = g()\n\t\tif err != nil {\n\t\t\tbreak\n\t\t}\n\t}\n\n\treturn err\n}\n", want: nil},
		{name: "goto passes", code: "package p\n\nfunc f() (err error) {\n\terr = g()\n\tif err != nil {\n\t\tgoto fail\n\t}\n\nfail:\n\treturn err\n}\n", want: nil},
		{name: "naked return of a named error result passes", code: "package p\n\nfunc f() (err error) {\n\terr = g()\n\tif err != nil {\n\t\treturn\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "named error result returned with values is reported", code: "package p\n\nfunc f() (n int, err error) {\n\tn, err = g()\n\tif err != nil {\n\t\treturn 0, nil\n\t}\n\n\treturn n, nil\n}\n", want: []string{"R26 p.go:5"}},
		{name: "error parameter the caller still owns passes", code: "package p\n\nfunc f(\n\terr error,\n) {\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "empty block passes", code: "package p\n\nfunc f() {\n\tif err := g(); err != nil {\n\t}\n}\n", want: nil},
		{name: "equality with nil passes", code: "package p\n\nfunc f() {\n\tif err := g(); err == nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "non-error pointer checked against nil passes", code: "package p\n\nfunc f() {\n\tif node := g(); node != nil {\n\t\treturn\n\t}\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
