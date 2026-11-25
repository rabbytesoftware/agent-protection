package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
)

func TestDroppederrorFailureIdentification(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "name ending in Err is an error", code: "package p\n\nfunc f() {\n\tif readErr := g(); readErr != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "variable declared as error is an error", code: "package p\n\nfunc f() {\n\tvar problem error = g()\n\tif problem != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:5"}},
		{name: "sole result of a same-package erroring function is an error", code: "package p\n\nfunc save() error\n\nfunc f() {\n\tproblem := save()\n\tif problem != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:7"}},
		{name: "last result of a same-package erroring method is an error", code: "package p\n\nfunc (s store) load() (n int, failure error)\n\nfunc f() {\n\tn, problem := s.load()\n\tif problem != nil {\n\t\treturn\n\t}\n\n\tuse(n)\n}\n", want: []string{"R26 p.go:7"}},
		{name: "result of an erroring interface method is an error", code: "package p\n\ntype saver interface {\n\tSave() error\n}\n\nfunc f() {\n\tvar problem = s.Save()\n\tif problem != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:9"}},
		{name: "callback parameter declared as error is an error", code: "package p\n\nfunc f() {\n\twalk(func(problem error) {\n\t\tif problem != nil {\n\t\t\treturn\n\t\t}\n\t})\n}\n", want: []string{"R26 p.go:5"}},
		{name: "package-level function value is not an error", code: "package p\n\nfunc callback()\n\nfunc f() {\n\tif callback != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "result of an indexed call is not an error", code: "package p\n\nfunc f() {\n\tvalue := makers[0]()\n\tif value != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "earlier result of an erroring function is not an error", code: "package p\n\nfunc load() (*int, error)\n\nfunc f() {\n\tvalue, err := load()\n\tif value != nil {\n\t\treturn\n\t}\n\n\tuse(err)\n}\n", want: nil},
		{name: "result of a function that returns no error is not an error", code: "package p\n\nfunc find() *int\n\nfunc f() {\n\tvalue := find()\n\tif value != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "result of an unknown function is not an error", code: "package p\n\nfunc f() {\n\tvalue := other.Find()\n\tif value != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "result of a non-call is not an error", code: "package p\n\nfunc f() {\n\tvalue := cache[key]\n\tif value != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "unresolved identifier is matched by name", code: "package p\n\nfunc f() {\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "shadowed recover is not the builtin", code: "package p\n\nfunc recover() *int\n\nfunc f() {\n\tif r := recover(); r != nil {\n\t\treturn\n\t}\n}\n", want: nil},
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

func TestDroppederrorIgnoredParameter(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "blank error parameter is reported", path: "p.go", code: "package p\n\nfunc Silence(\n\t_ error,\n) {\n}\n", want: []string{"R26 p.go:4"}},
		{name: "unnamed error parameter is reported", path: "p.go", code: "package p\n\nfunc Silence(error) {\n}\n", want: []string{"R26 p.go:3"}},
		{name: "named error parameter never read is reported", path: "p.go", code: "package p\n\ntype quiet struct{}\n\nfunc (quiet) Report(\n\terr error,\n) {\n\tcount++\n}\n", want: []string{"R26 p.go:6"}},
		{name: "func literal ignoring its error parameter is reported", path: "p.go", code: "package p\n\nfunc f() {\n\twire(func(err error) {\n\t})\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error parameter read in the body passes", path: "p.go", code: "package p\n\nfunc Report(\n\terr error,\n) {\n\trecord(err)\n}\n", want: nil},
		{name: "error parameter shadowed and never read is reported", path: "p.go", code: "package p\n\nfunc Report(\n\terr error,\n) {\n\tif err := other(); err != nil {\n\t\trecord(err)\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "declaration without a body passes", path: "p.go", code: "package p\n\nfunc Report(\n\terr error,\n)\n", want: nil},
		{name: "non-error parameter never read passes", path: "p.go", code: "package p\n\nfunc Report(\n\tvalue int,\n) {\n}\n", want: nil},
		{name: "func literal default reporter is reported", path: "p.go", code: "package p\n\nfunc defaults(\n\tc config,\n) (completed config) {\n\tcompleted = c\n\tcompleted.report = func(error) {\n\t}\n\n\treturn completed\n}\n", want: []string{"R26 p.go:7"}},
		{name: "func literal passed as a reporter argument is reported", path: "p.go", code: "package p\n\nfunc f() {\n\tstart(func(_ error) {\n\t})\n}\n", want: []string{"R26 p.go:4"}},
		{name: "method with an unnamed error parameter is reported", path: "p.go", code: "package p\n\ntype sink struct{}\n\nfunc (s sink) Report(error) {\n}\n", want: []string{"R26 p.go:5"}},
		{name: "variadic error parameter never read is reported", path: "p.go", code: "package p\n\nfunc Report(\n\terrs ...error,\n) {\n}\n", want: []string{"R26 p.go:4"}},
		{name: "one of two error parameters never read is reported", path: "p.go", code: "package p\n\nfunc Report(\n\tfirst, second error,\n) {\n\trecord(first)\n}\n", want: []string{"R26 p.go:4"}},
		{name: "test helpers ignoring an error pass", path: "p_test.go", code: "package p\n\nfunc ignore(\n\t_ error,\n) {\n}\n", want: nil},
		{name: "mock files ignoring an error pass", path: "mocks/p.go", code: "package mocks\n\nfunc ignore(\n\t_ error,\n) {\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{parsedAt(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}
