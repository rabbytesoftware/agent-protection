package entrypoint_test

import "testing"

func TestEntrypointErrorGuards(
	t *testing.T,
) {
	assertBodies(
		t,
		[]bodyCase{
			{name: "error guard with init passes", body: "if err := impl.Check(); err != nil {\nreturn nil, err\n}\nreturn nil, nil", want: nil},
			{name: "reversed error guard passes", body: "err := impl.Check()\nif nil != err {\nreturn nil, err\n}\nreturn nil, nil", want: nil},
			{name: "guard on an Err suffixed variable passes", body: "openErr := impl.Check()\nif openErr != nil {\nreturn nil, openErr\n}\nreturn nil, nil", want: nil},
			{name: "guard on a call's last result passes", body: "q, failure := build()\nif failure != nil {\nreturn nil, failure\n}\nreturn q, nil", want: nil},
			{name: "guard on a var declared from a call passes", body: "var q, failure = build()\nif failure != nil {\nreturn nil, failure\n}\nreturn q, nil", want: nil},
			{name: "nil check on a parameter is reported", body: "if cached != nil {\nreturn cached, nil\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "nil check on a single call result is reported", body: "q := impl.Lookup()\nif q != nil {\nreturn nil, nil\n}\nreturn nil, nil", want: []string{"R14 container.go:24"}},
			{name: "equality check is reported", body: "err := impl.Check()\nif err == nil {\nreturn nil, err\n}\nreturn nil, nil", want: []string{"R14 container.go:24"}},
			{name: "branch that is not an error guard is reported", body: "if impl.Ready() {\nreturn nil, nil\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "guard with an else is reported", body: "if err := impl.Check(); err != nil {\nreturn nil, err\n} else {\nimpl.Check()\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "guard doing more than return is reported", body: "if err := impl.Check(); err != nil {\nimpl.Check()\nreturn nil, err\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "guard on a selector is reported", body: "if q.err != nil {\nreturn nil, nil\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
		},
	)
}

func TestEntrypointGuardsKeepTheirReturnShape(
	t *testing.T,
) {
	assertFindings(
		t,
		"cached parameter guard from the red team",
		checked(
			t,
			map[string]string{
				"container.go": "package arith\n\nimport \"example.com/m/internal/cube\"\n\ntype Calculator interface{ Cube(int) int }\n\nfunc Override(\n\tcached Calculator,\n) (calculator Calculator) {\n\tif cached != nil {\n\t\treturn cached\n\t}\n\n\treturn cube.Cube()\n}\n",
			},
		),
		[]string{"R14 container.go:10"},
	)
}
