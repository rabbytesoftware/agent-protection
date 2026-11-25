package publicapi_test

import (
	"strings"
	"testing"
)

func TestPublicapiEntryFunctionsOnlyDelegate(
	t *testing.T,
) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{name: "calls, assignments and returns pass", body: "a, err := build()\n_ = a\nreturn &Q{V: a.V}, err", want: nil},
		{name: "error guard with init passes", body: "if err := check(); err != nil {\nreturn nil, err\n}\nreturn nil, nil", want: nil},
		{name: "reversed error guard passes", body: "err := check()\nif nil != err {\nreturn nil, err\n}\nreturn nil, nil", want: nil},
		{name: "for loop is reported once", body: "for i := 0; i < 1; i++ {\nfor {\n}\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "range loop is reported", body: "for range items() {\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "switch is reported", body: "switch {\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "type switch is reported", body: "switch any(nil).(type) {\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "select is reported", body: "select {\n}", want: []string{"R14 container.go:6"}},
		{name: "goroutine is reported", body: "go check()\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "defer is reported", body: "defer check()\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "func literal is reported", body: "run(func() {})\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "branch that is not an error guard is reported", body: "if ready() {\nreturn nil, nil\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "guard with an else is reported", body: "if err := check(); err != nil {\nreturn nil, err\n} else {\ncheck()\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "guard doing more than return is reported", body: "if err := check(); err != nil {\ncheck()\nreturn nil, err\n}\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "arithmetic is reported", body: "v := 1 + size()\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "negation is reported", body: "v := !ready()\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "minus is reported", body: "v := -size()\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:6"}},
		{name: "compound assignment is reported", body: "v := size()\nv += 1\nreturn nil, nil", want: []string{"R14 container.go:7"}},
		{name: "increment is reported", body: "v := size()\nv++\nreturn nil, nil", want: []string{"R14 container.go:7"}},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				map[string]string{"container.go": entryFunction(tc.body)},
			),
			tc.want,
		)
	}
}

func TestPublicapiEntryFunctionWithoutBodyPasses(
	t *testing.T,
) {
	assertFindings(
		t,
		"entry function without a body passes",
		checked(
			t,
			map[string]string{"container.go": "package jobq\n\nfunc New() int\n"},
		),
		nil,
	)
}

func TestPublicapiDelegationSnippet(
	t *testing.T,
) {
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": entryFunction("for {\n}")},
		),
		"container.go:6",
		"container.go func New loops: entry functions only delegate; move the logic into an internal package",
	)
}

func entryFunction(
	body string,
) (code string) {
	return "package jobq\n\ntype Q struct{ V int }\n\nfunc New() (*Q, error) {\n" + strings.TrimSpace(body) + "\n}\n"
}
