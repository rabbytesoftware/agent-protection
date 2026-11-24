package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDroppederrorDiscarded(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "error variable assigned to blank is reported", code: "package p\n\nfunc f() {\n\terr := g()\n\t_ = err\n}\n", want: []string{"R26 p.go:5"}},
		{name: "recover result assigned to blank is reported", code: "package p\n\nfunc f() {\n\t_ = recover()\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error among paired blank assignments is reported at the error", code: "package p\n\nfunc f() {\n\t_, _ = n,\n\t\terr\n}\n", want: []string{"R26 p.go:5"}},
		{name: "non-error value assigned to blank passes", code: "package p\n\nfunc f() {\n\t_ = value\n}\n", want: nil},
		{name: "error assigned to a named variable passes", code: "package p\n\nfunc f() {\n\tlast = err\n}\n", want: nil},
		{name: "blank result of a multi-value call is left to errcheck", code: "package p\n\nfunc f() {\n\tn, _ := g()\n\tuse(n)\n}\n", want: nil},
		{name: "bare recover statement is reported", code: "package p\n\nfunc f() {\n\tdefer func() {\n\t\trecover()\n\t}()\n}\n", want: []string{"R26 p.go:5"}},
		{name: "recover compared with nil is reported", code: "package p\n\nfunc f() bool {\n\treturn recover() != nil\n}\n", want: []string{"R26 p.go:4"}},
		{name: "recover expected to be nil passes", code: "package p\n\nfunc f() {\n\tif recover() == nil {\n\t\tfail()\n\t}\n}\n", want: nil},
		{name: "other comparisons with recover pass", code: "package p\n\nfunc f() bool {\n\treturn recover() == sentinel\n}\n", want: nil},
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
