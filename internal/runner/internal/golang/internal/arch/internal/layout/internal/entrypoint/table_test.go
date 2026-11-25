package entrypoint_test

import "testing"

func TestEntrypointTablesAndIndexes(
	t *testing.T,
) {
	assertBodies(
		t,
		[]bodyCase{
			{name: "map of interfaces declared here is reported", body: "impl.Use(map[string]Runner{\"a\": impl.A()})\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "slice of funcs is reported", body: "impl.Use([]func(){impl.A})\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "slice of any is reported", body: "impl.Use([]any{impl.A()})\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "map of interface literals is reported", body: "impl.Use(map[string]interface{ Run() }{})\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "slice of plain values passes", body: "impl.Use([]int{1, 2})\nreturn nil, nil", want: nil},
			{name: "slice of another package's structs passes", body: "impl.Use([]impl.Checker{{Name: \"a\"}})\nreturn nil, nil", want: nil},
			{name: "struct literal passes", body: "return &Q{V: size}, nil", want: nil},
			{name: "index into a local is reported", body: "items := impl.Items()\nreturn items[0], nil", want: []string{"R14 container.go:24"}},
			{name: "index into a literal is reported", body: "return map[string]*Q{}[\"a\"], nil", want: []string{"R14 container.go:23"}},
			{name: "calling an indexed func is reported", body: "handlers := impl.Handlers()\nhandlers[0]()\nreturn nil, nil", want: []string{"R14 container.go:24"}},
			{name: "generic instantiation of an own function passes", body: "return generic[int](), nil", want: nil},
			{name: "generic type in a composite literal passes", body: "impl.Use(impl.Box[int]{})\nreturn nil, nil", want: nil},
			{name: "generic type in a var declaration passes", body: "var box impl.Box[int]\nimpl.Use(box)\nreturn nil, nil", want: nil},
			{name: "generic type in a type assertion passes", body: "impl.Use(impl.Any().(impl.Box[int]))\nreturn nil, nil", want: nil},
		},
	)
}

func TestEntrypointTableSnippet(
	t *testing.T,
) {
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": entry("impl.Use([]any{impl.A()})\nreturn nil, nil")},
		),
		"container.go:23",
		"container.go func New builds a table of funcs or interfaces: entry functions only delegate; move the logic into an internal package",
	)
}

func TestEntrypointHiddenSelectionInContainer(
	t *testing.T,
) {
	assertFindings(
		t,
		"table, index, configuration lookups and a non-error guard in container.go",
		checked(
			t,
			map[string]string{
				"container.go":  "package arith\n\nimport (\n\t\"cmp\"\n\t\"os\"\n\t\"strings\"\n\n\t\"example.com/m/internal/cube\"\n\t\"example.com/m/internal/square\"\n)\n\nfunc Arith() (calculator Calculator) {\n\treturn map[string]Calculator{\n\t\t\"square\": cube.Cube(square.Square()),\n\t\t\"cached\": cube.Cube(cube.Cube(square.Square())),\n\t}[strings.ToLower(cmp.Or(\n\t\tos.Getenv(\"ARITH_MODE\"),\n\t\t\"square\",\n\t))]\n}\n\nfunc Override(\n\tcached Calculator,\n) (calculator Calculator) {\n\tif cached != nil {\n\t\treturn cached\n\t}\n\n\treturn cube.Cube(cube.Cube(square.Square()))\n}\n",
				"calculator.go": "package arith\n\ntype Calculator interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n\tCube(\n\t\tvalue int,\n\t) (result int)\n}\n",
			},
		),
		[]string{
			"R14 container.go:13",
			"R14 container.go:13",
			"R14 container.go:16",
			"R14 container.go:16",
			"R14 container.go:17",
			"R14 container.go:25",
		},
	)
}
