package entrypoint_test

import "testing"

func TestEntrypointResults(
	t *testing.T,
) {
	cases := []struct {
		name    string
		results string
		want    []string
	}{
		{name: "own interface passes", results: "(calculator Calculator)", want: nil},
		{name: "own type pointer passes", results: "(*Result, error)", want: nil},
		{name: "basic types and error pass", results: "(int, string, error)", want: nil},
		{name: "slices and maps of own types pass", results: "([]Calculator, map[string]Result)", want: nil},
		{name: "value type of another package passes", results: "(report.Finding)", want: nil},
		{name: "slice of another package's value type passes", results: "([]report.Finding, error)", want: nil},
		{name: "interface literal passes", results: "(interface{ Run() })", want: nil},
		{name: "type parameter passes", results: "(T)", want: nil},
		{name: "generic own type passes", results: "(Box[int])", want: nil},
		{name: "internal struct is reported", results: "(calculator cube.Cube)", want: []string{"R64 container.go:17"}},
		{name: "pointer to internal struct is reported", results: "(*cube.Cube, error)", want: []string{"R64 container.go:17"}},
		{name: "slice of internal struct is reported", results: "([]cube.Cube)", want: []string{"R64 container.go:17"}},
		{name: "map of internal struct is reported", results: "(map[string]cube.Cube)", want: []string{"R64 container.go:17"}},
		{name: "channel of internal struct is reported", results: "(<-chan cube.Cube)", want: []string{"R64 container.go:17"}},
		{name: "generic own type over an internal struct is reported", results: "(Box[cube.Cube])", want: []string{"R64 container.go:17"}},
		{name: "internal generic type is reported", results: "(cube.Pair[int, string])", want: []string{"R64 container.go:17"}},
		{name: "pointer into another package is reported", results: "(*report.Finding)", want: []string{"R64 container.go:17"}},
		{name: "pointer to a generic type of another package is reported", results: "(*report.Box[int])", want: []string{"R64 container.go:17"}},
		{name: "func result is reported", results: "(func())", want: []string{"R64 container.go:17"}},
		{name: "struct literal result is reported", results: "(struct{ V int })", want: []string{"R64 container.go:17"}},
		{name: "unknown identifier is reported", results: "(Hidden)", want: []string{"R64 container.go:17"}},
		{name: "one finding per offending result", results: "(cube.Cube, *cube.Cube, error)", want: []string{"R64 container.go:17", "R64 container.go:17"}},
		{name: "selector on a non-import passes", results: "(x.Cube)", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				map[string]string{"container.go": resultEntry(tc.results)},
			),
			tc.want,
		)
	}
}

func TestEntrypointResultsWithoutResults(
	t *testing.T,
) {
	assertFindings(
		t,
		"entry without results passes",
		checked(
			t,
			map[string]string{"container.go": "package arith\n\nfunc Start() {\n}\n"},
		),
		nil,
	)
}

func TestEntrypointResultSnippet(
	t *testing.T,
) {
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"container.go": resultEntry("(calculator cube.Cube)")},
		),
		"container.go:17",
		"container.go func Arith returns cube.Cube: entry functions return an interface or type declared in package arith, a basic type or error, never a type from an internal package or a pointer into another package; declare the interface callers need in . (or alias a declaration-only internal type there) and return the implementation through it",
	)
}

func resultEntry(
	results string,
) (code string) {
	return "package arith\n\nimport (\n\t\"example.com/m/internal/cube\"\n\t\"example.com/m/report\"\n)\n\ntype Calculator interface{ Cube(int) int }\n\ntype Result struct{}\n\ntype Box[T any] struct{ V T }\n\nvar _ report.Finding\n\nfunc Arith[T any](\n) " + results + " {\n\treturn cube.Cube{}\n}\n"
}

func TestEntrypointConcreteInternalResult(
	t *testing.T,
) {
	assertFindings(
		t,
		"container entry returning a concrete internal struct",
		checked(
			t,
			map[string]string{"container.go": "package arith\n\nimport (\n\t\"example.com/m/internal/cube\"\n\t\"example.com/m/internal/square\"\n)\n\nfunc Arith() (calculator cube.Cube) {\n\treturn cube.Cube{Squarer: square.Square(), Memo: map[int]int{}}\n}\n"},
		),
		[]string{"R64 container.go:8"},
	)
}
