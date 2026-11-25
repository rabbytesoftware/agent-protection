package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
)

func TestDroppederrorCalleesSpanTheDirectory(
	t *testing.T,
) {
	cases := []struct {
		name       string
		callerPath string
		want       []string
	}{
		{name: "erroring function declared in a sibling file is known", callerPath: "pkg/b.go", want: []string{"R26 pkg/b.go:5"}},
		{name: "erroring function declared in another directory is unknown", callerPath: "other/b.go", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{
				parsedAt(
					t,
					"pkg/a.go",
					"package p\n\nfunc save() (failure error)\n",
				),
				parsedAt(
					t,
					tc.callerPath,
					"package p\n\nfunc f() {\n\tproblem := save()\n\tif problem != nil {\n\t\treturn\n\t}\n}\n",
				),
			}),
			tc.want,
		)
	}
}

func TestDroppederrorKnowsStandardAndModuleCallees(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "blank error of strconv.Atoi is reported", code: "package p\n\nimport \"strconv\"\n\nfunc f(s string) (n int) {\n\tn, _ = strconv.Atoi(s)\n\treturn n\n}\n", want: []string{"R26 pkg/b.go:6"}},
		{name: "blank error of an aliased encoding/json call is reported", code: "package p\n\nimport js \"encoding/json\"\n\nfunc f(b []byte) {\n\t_ = js.Unmarshal(b, nil)\n}\n", want: []string{"R26 pkg/b.go:6"}},
		{name: "blank error of fmt.Sscanf is reported", code: "package p\n\nimport \"fmt\"\n\nfunc f(s string) (n int) {\n\t_, _ = fmt.Sscanf(s, \"%d\", &n)\n\treturn n\n}\n", want: []string{"R26 pkg/b.go:6"}},
		{name: "a stdlib error checked and dropped is reported", code: "package p\n\nimport \"strconv\"\n\nfunc f(s string) (n int) {\n\tn, e := strconv.Atoi(s)\n\tif e != nil {\n\t\treturn 0\n\t}\n\n\treturn n\n}\n", want: []string{"R26 pkg/b.go:7"}},
		{name: "blank error of a module package function is reported", code: "package p\n\nimport \"example.com/m/store\"\n\nfunc f() {\n\t_ = store.Save()\n}\n", want: []string{"R26 pkg/b.go:6"}},
		{name: "a blank value of a non-erroring stdlib call passes", code: "package p\n\nimport \"strconv\"\n\nfunc f(n int) {\n\t_ = strconv.Itoa(n)\n}\n", want: nil},
		{name: "an unknown package passes", code: "package p\n\nimport \"example.com/other/parse\"\n\nfunc f(s string) (n int) {\n\tn, _ = parse.Atoi(s)\n\treturn n\n}\n", want: nil},
		{name: "a method on a value named like a package passes", code: "package p\n\nfunc f(strconv parser, s string) (n int) {\n\tn, _ = strconv.Atoi(s)\n\treturn n\n}\n", want: nil},
		{name: "blank and dot imports add no callees", code: "package p\n\nimport (\n\t_ \"os\"\n\t. \"strconv\"\n)\n\nfunc f(s string) (n int) {\n\tn, _ = Atoi(s)\n\treturn n\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{
				parsedAt(
					t,
					"store/store.go",
					"package store\n\nfunc Save() (failure error)\n",
				),
				parsedAt(
					t,
					"pkg/b.go",
					tc.code,
				),
			}),
			tc.want,
		)
	}
}
