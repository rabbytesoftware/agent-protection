package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
)

func TestDroppederrorOccurrences(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "use of an earlier value does not cover a reassigned one", code: "package p\n\nfunc f() {\n\terr := g()\n\tuse(err)\n\terr = g()\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:7"}},
		{name: "reassignment after the check does not hide the drop", code: "package p\n\nfunc f() {\n\terr := g()\n\tif err != nil {\n\t\treturn\n\t}\n\n\terr = g()\n\tuse(err)\n}\n", want: []string{"R26 p.go:5"}},
		{name: "range-bound error skipped with continue is reported", code: "package p\n\nfunc f() {\n\tfor _, err := range errs {\n\t\tif err != nil {\n\t\t\tcontinue\n\t\t}\n\t}\n}\n", want: []string{"R26 p.go:5"}},
		{name: "declared error assigned later and dropped is reported", code: "package p\n\nfunc f() {\n\tvar err error\n\terr = g()\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:6"}},
		{name: "same-named error of another function does not count as a use", code: "package p\n\nfunc f() {\n\tif err := g(); err != nil {\n\t\treturn\n\t}\n}\n\nfunc h() error {\n\terr := g()\n\treturn err\n}\n", want: []string{"R26 p.go:4"}},
		{name: "unresolved error used through a same-named field passes", code: "package p\n\nfunc f() {\n\tif err != nil {\n\t\ts.err = nil\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "use inside the else branch passes", code: "package p\n\nfunc f() {\n\tif err := g(); err != nil {\n\t\treturn\n\t} else {\n\t\tuse(err)\n\t}\n}\n", want: nil},
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
