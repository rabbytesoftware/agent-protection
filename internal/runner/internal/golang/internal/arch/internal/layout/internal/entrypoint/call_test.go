package entrypoint_test

import "testing"

func TestEntrypointCalls(
	t *testing.T,
) {
	assertBodies(
		t,
		[]bodyCase{
			{name: "own internal package call passes", body: "return impl.New(size), nil", want: nil},
			{name: "own declaration call passes", body: "return build()", want: nil},
			{name: "other module constructor passes", body: "impl.Use(lib.New())\nreturn nil, nil", want: nil},
			{name: "method call on a local passes", body: "q, err := build()\nq.Start()\nreturn q, err", want: nil},
			{name: "func parameter call passes", body: "impl.Use(size)\nreturn nil, nil", want: nil},
			{name: "error constructors pass", body: "return nil, errors.Join(errors.New(\"a\"), fmt.Errorf(\"b: %w\", impl.Err))", want: nil},
			{name: "slices.Concat passes", body: "impl.Use(slices.Concat(impl.A(), impl.B()))\nreturn nil, nil", want: nil},
			{name: "generic constructor passes", body: "impl.Use(impl.New[int]())\nreturn nil, nil", want: nil},
			{name: "environment lookup is reported", body: "impl.Use(os.Getenv(\"MODE\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "strings call is reported", body: "impl.Use(strings.ToLower(\"A\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "os.Exit is reported outside main", body: "os.Exit(1)\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "builtin call is reported", body: "impl.Use(len(\"a\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "conversion to a composite type is reported", body: "impl.Use([]byte(\"a\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "shadowed qualifier is a method call", body: "os := impl.Open()\nos.Getenv()\nreturn nil, nil", want: nil},
		},
	)
}

func TestEntrypointCallsInModulesWithoutDomains(
	t *testing.T,
) {
	assertFindings(
		t,
		"own packages of a module path without a dot are not standard library",
		checkedAt(
			t,
			"ledger",
			map[string]string{"container.go": "package ledger\n\nimport \"ledger/internal/book\"\n\nfunc New() int {\n\treturn book.New()\n}\n"},
		),
		nil,
	)
}
