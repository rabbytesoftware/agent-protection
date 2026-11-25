package entrypoint_test

import (
	"strings"
	"testing"
)

type bodyCase struct {
	name string
	body string
	want []string
}

func TestEntrypointEntryFunctionsOnlyDelegate(
	t *testing.T,
) {
	assertBodies(
		t,
		[]bodyCase{
			{name: "calls, assignments and returns pass", body: "a, err := build()\n_ = a\nreturn &Q{V: a.V}, err", want: nil},
			{name: "for loop is reported once", body: "for i := 0; i < 1; i++ {\nfor {\n}\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "range loop is reported", body: "for range impl.Items() {\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "switch is reported", body: "switch {\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "type switch is reported", body: "switch any(nil).(type) {\n}\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "select is reported", body: "select {\n}", want: []string{"R14 container.go:23"}},
			{name: "goroutine is reported", body: "go build()\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "defer is reported", body: "defer build()\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "func literal is reported", body: "impl.Run(func() {})\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "arithmetic is reported", body: "v := 1 + size\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "negation is reported", body: "v := !impl.Ready()\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "minus is reported", body: "v := -size\n_ = v\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "compound assignment is reported", body: "v := size\nv += 1\nreturn nil, nil", want: []string{"R14 container.go:24"}},
			{name: "increment is reported", body: "v := size\nv++\nreturn nil, nil", want: []string{"R14 container.go:24"}},
			{name: "address of a composite literal passes", body: "return &Q{V: size}, nil", want: nil},
			{name: "var declaration passes", body: "var q *Q\nreturn q, nil", want: nil},
			{name: "setting a field of a local by hand is reported", body: "q := &Q{}\nq.V = size\nreturn q, nil", want: []string{"R14 container.go:24"}},
			{name: "setting a method value into a field is reported once per statement", body: "q, err := build()\nq.V, q.W = q.Run, q.Stop\nreturn q, err", want: []string{"R14 container.go:24"}},
			{name: "assigning a package field is left alone", body: "impl.Default = size\nreturn nil, nil", want: nil},
		},
	)
}

func TestEntrypointEntryFunctionWithoutBodyPasses(
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

func TestEntrypointDelegationSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{"container.go": entry("for {\n}")},
	)
	assertSnippet(
		t,
		findings,
		"container.go:23",
		"container.go func New loops: entry functions only delegate; move the logic into an internal package",
	)
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": entry("impl.Set(os.Getenv(\"MODE\"))\nreturn nil, nil")},
		),
		"container.go:23",
		"container.go func New calls os.Getenv: entry functions only delegate; call only this module's internal packages, its own declarations, errors.New, errors.Join, fmt.Errorf, slices.Concat or other modules' constructors, and take configuration as parameters",
	)
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": entry("q := &Q{}\nq.V = size\nreturn q, nil")},
		),
		"container.go:24",
		"container.go func New sets field q.V by hand: entry functions only delegate; a container passes dependencies to constructors and returns what they build; construct the object in an internal package and return it as this module's interface instead of filling its fields here",
	)
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": entry("lib.Init()\nreturn nil, nil")},
		),
		"container.go:23",
		"container.go func New calls lib.Init without wiring its result: entry functions only delegate; call another module's constructor only to assign its result and pass it on or return it; flags, environment and configuration files are read by the caller and arrive as entry parameters",
	)
}

func assertBodies(
	t *testing.T,
	cases []bodyCase,
) {
	t.Helper()
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				map[string]string{"container.go": entry(tc.body)},
			),
			tc.want,
		)
	}
}

func entry(
	body string,
) (code string) {
	return "package jobq\n\nimport (\n\t\"errors\"\n\t\"fmt\"\n\t\"os\"\n\t\"slices\"\n\t\"strings\"\n\n\t\"example.com/m/internal/impl\"\n\t\"github.com/other/lib\"\n)\n\ntype Q struct{ V int }\n\ntype Runner interface{ Run() }\n\nfunc build() (*Q, error) { return nil, nil }\n\nfunc generic[T any]() *Q { return nil }\n\nfunc New(size int) (*Q, error) {\n" + strings.TrimSpace(body) + "\n}\n"
}
