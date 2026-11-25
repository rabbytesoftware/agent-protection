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
