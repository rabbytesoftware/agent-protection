package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
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
		{name: "erroring call assigned to blank is reported", code: "package p\n\nfunc save() error\n\nfunc f() {\n\t_ = save()\n}\n", want: []string{"R26 p.go:6"}},
		{name: "every result of an erroring call assigned to blank is reported", code: "package p\n\nfunc load() (int, error)\n\nfunc f() {\n\t_, _ = load()\n}\n", want: []string{"R26 p.go:6"}},
		{name: "error result of an erroring call assigned to blank is reported", code: "package p\n\nfunc load() (int, error)\n\nfunc f() int {\n\tn, _ := load()\n\treturn n\n}\n", want: []string{"R26 p.go:6"}},
		{name: "blank value result of an erroring call passes", code: "package p\n\nfunc load() (int, error)\n\nfunc f() error {\n\t_, err := load()\n\treturn err\n}\n", want: nil},
		{name: "non-erroring call assigned to blank passes", code: "package p\n\nfunc size() int\n\nfunc f() {\n\t_ = size()\n}\n", want: nil},
		{name: "bare recover statement is reported", code: "package p\n\nfunc f() {\n\tdefer func() {\n\t\trecover()\n\t}()\n}\n", want: []string{"R26 p.go:5"}},
		{name: "recover compared with nil is reported", code: "package p\n\nfunc f() bool {\n\treturn recover() != nil\n}\n", want: []string{"R26 p.go:4"}},
		{name: "recover expected to be nil passes", code: "package p\n\nfunc f() {\n\tif recover() == nil {\n\t\tfail()\n\t}\n}\n", want: nil},
		{name: "other comparisons with recover pass", code: "package p\n\nfunc f() bool {\n\treturn recover() == sentinel\n}\n", want: nil},
		{name: "known erroring call checked inline and returned from is reported", code: "package p\n\nimport \"os\"\n\nfunc f(\n\tp string,\n) {\n\tif os.Remove(p) != nil {\n\t\treturn\n\t}\n\n\tuse(p)\n}\n", want: []string{"R26 p.go:8"}},
		{name: "module erroring call checked inline and continued past is reported", code: "package p\n\nfunc save() error\n\nfunc f() {\n\tfor range 3 {\n\t\tif save() != nil {\n\t\t\tcontinue\n\t\t}\n\t}\n}\n", want: []string{"R26 p.go:7"}},
		{name: "known erroring call reduced to a bool is reported", code: "package p\n\nimport \"os\"\n\nfunc f(\n\tp string,\n) (ok bool) {\n\tok = os.Remove(p) == nil\n\treturn ok\n}\n", want: []string{"R26 p.go:8"}},
		{name: "known erroring call expected to succeed in production is reported", code: "package p\n\nimport \"os\"\n\nfunc f(\n\tp string,\n) {\n\tif os.Remove(p) == nil {\n\t\tuse(p)\n\t}\n}\n", want: []string{"R26 p.go:8"}},
		{name: "inline check whose body handles the failure passes", code: "package p\n\nimport \"os\"\n\nfunc f(\n\tp string,\n) {\n\tif os.Remove(p) != nil {\n\t\tfail(p)\n\t}\n}\n", want: nil},
		{name: "unknown call compared with nil passes", code: "package p\n\nfunc f() {\n\tif lookup() != nil {\n\t\treturn\n\t}\n}\n", want: nil},
		{name: "call with a value and an error result compared with nil passes", code: "package p\n\nimport \"os\"\n\nfunc f() (ok bool) {\n\treturn os.Getwd != nil\n}\n", want: nil},
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

func TestDroppederrorInlineInTests(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "test expecting an error inline passes", code: "package p\n\nimport \"encoding/json\"\n\nfunc TestF(t *testing.T) {\n\tif json.Unmarshal(nil, &v) == nil {\n\t\tt.Fatal(\"want error\")\n\t}\n}\n", want: nil},
		{name: "test comparing an inline error outcome with a want passes", code: "package p\n\nfunc save() error\n\nfunc TestF(t *testing.T) {\n\tfailed := save() != nil\n\tcheck(t, failed)\n}\n", want: nil},
		{name: "test returning on an inline error is reported", code: "package p\n\nfunc save() error\n\nfunc TestF(t *testing.T) {\n\tif save() != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p_test.go:6"}},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{parsedAt(
				t,
				"p_test.go",
				tc.code,
			)}),
			tc.want,
		)
	}
}
