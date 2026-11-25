package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
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
		{name: "error classified with errors.Is and dropped otherwise is reported", code: "package p\n\nfunc f() error {\n\tif err := g(); err != nil && errors.Is(err, ErrMissing) {\n\t\treturn nil\n\t}\n\n\treturn nil\n}\n", want: []string{"R26 p.go:4"}},
		{name: "partial errors.Is check with nothing after is reported", code: "package p\n\nfunc f(\n\ttext string,\n) (value int) {\n\tvalue, err := strconv.Atoi(text)\n\tif errors.Is(\n\t\terr,\n\t\tstrconv.ErrSyntax,\n\t) {\n\t\treturn 0\n\t}\n\n\treturn value\n}\n", want: []string{"R26 p.go:7"}},
		{name: "partial errors.As check wrapping only the match is reported", code: "package p\n\nfunc f() error {\n\terr := g()\n\tvar target *E\n\tif errors.As(err, &target) {\n\t\treturn fmt.Errorf(\"g: %w\", err)\n\t}\n\n\treturn nil\n}\n", want: []string{"R26 p.go:6"}},
		{name: "chained matches with nothing after report the last check", code: "package p\n\nfunc f() int {\n\terr := g()\n\tif errors.Is(err, ErrA) {\n\t\treturn 1\n\t} else if errors.Is(err, ErrB) {\n\t\treturn 2\n\t}\n\n\treturn 0\n}\n", want: []string{"R26 p.go:7"}},
		{name: "errors.Is followed by a nil check passes", code: "package p\n\nfunc f() error {\n\terr := g()\n\tif errors.Is(err, os.ErrNotExist) {\n\t\treturn nil\n\t}\n\n\tif err != nil {\n\t\treturn fmt.Errorf(\"g: %w\", err)\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "errors.Is with the error returned in the else passes", code: "package p\n\nfunc f() error {\n\tif err := g(); errors.Is(err, ErrMissing) {\n\t\treturn nil\n\t} else {\n\t\treturn err\n\t}\n}\n", want: nil},
		{name: "negated errors.Is returning the error passes", code: "package p\n\nfunc f() error {\n\tif err := g(); err != nil && !errors.Is(err, ErrMissing) {\n\t\treturn err\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "negated errors.Is reporting a failure passes", code: "package p\n\nfunc f() {\n\t_, err := os.Stat(path)\n\tif !errors.Is(err, fs.ErrNotExist) {\n\t\tfail(\"left behind\")\n\t}\n}\n", want: nil},
		{name: "negated errors.Is falling back silently is reported", code: "package p\n\nfunc f() int {\n\t_, err := os.Stat(path)\n\tif !errors.Is(err, fs.ErrNotExist) {\n\t\treturn 0\n\t}\n\n\treturn 1\n}\n", want: []string{"R26 p.go:5"}},
		{name: "errors.Is followed by another check in the condition passes", code: "package p\n\nfunc f() error {\n\terr := g()\n\tif errors.Is(err, ErrMissing) || err == nil {\n\t\treturn nil\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "use after a rebinding does not cover the match", code: "package p\n\nfunc f() error {\n\terr := g()\n\tif errors.Is(err, ErrMissing) {\n\t\treturn nil\n\t}\n\n\terr = h()\n\treturn err\n}\n", want: []string{"R26 p.go:5"}},
		{name: "errors.Is on an error parameter passes", code: "package p\n\nfunc f(\n\terr error,\n) (missing bool) {\n\tif errors.Is(err, ErrMissing) {\n\t\treturn true\n\t}\n\n\treturn false\n}\n", want: nil},
		{name: "errors.Is on a non-error value passes", code: "package p\n\nfunc f() int {\n\tvalue := g()\n\tif errors.Is(value, ErrMissing) {\n\t\treturn 1\n\t}\n\n\treturn 0\n}\n", want: nil},
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
