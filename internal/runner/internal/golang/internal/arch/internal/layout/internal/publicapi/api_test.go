package publicapi_test

import "testing"

const apiSurface = `package jobq

import "errors"

type Queue interface {
	Enqueue() error
}

type Status int

type Failure struct{}

const Pending Status = 1

var ErrClosed = errors.New("closed")

var _ Queue = nil

func (f Failure) Error() string {
	return "failure"
}

func (s Status) String() string {
	return "status"
}
`

func TestPublicapiAPIShape(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "exported types, constants, Err sentinels, blanks and error methods pass",
				files: map[string]string{"jobq.go": apiSurface},
				want:  nil,
			},
			{
				name:  "unexported constant is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nconst (\n\tLimit = 1\n\tdefaultLimit = 2\n)\n"},
				want:  []string{"R14 jobq.go:5"},
			},
			{
				name:  "exported var that is not an Err sentinel is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nvar Default = 1\n"},
				want:  []string{"R14 jobq.go:3"},
			},
			{
				name:  "unexported type is reported",
				files: map[string]string{"jobq.go": "package jobq\n\ntype state struct{}\n"},
				want:  []string{"R14 jobq.go:3"},
			},
			{
				name:  "error method on an unexported receiver is reported",
				files: map[string]string{"jobq.go": "package jobq\n\ntype failure struct{}\n\nfunc (f failure) Error() string {\n\treturn \"\"\n}\n"},
				want:  []string{"R14 jobq.go:3", "R14 jobq.go:5"},
			},
			{
				name:  "init is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nfunc init() {}\n"},
				want:  []string{"R14 jobq.go:3"},
			},
			{
				name:  "exported function outside container.go is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nfunc New() int {\n\treturn 1\n}\n"},
				want:  []string{"R14 jobq.go:3"},
			},
			{
				name:  "delegating entry function in container.go passes",
				files: map[string]string{"container.go": "package jobq\n\nfunc New() int {\n\treturn build()\n}\n"},
				want:  nil,
			},
			{
				name:  "unexported function in container.go is reported",
				files: map[string]string{"container.go": "package jobq\n\nfunc New() int {\n\treturn build()\n}\n\nfunc build() int {\n\treturn 1\n}\n"},
				want:  []string{"R14 container.go:7"},
			},
		},
	)
}

func TestPublicapiAPISnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{"jobq.go": "package jobq\n\nfunc New() int {\n\treturn 1\n}\n\nvar cache = 1\n"},
	)
	assertSnippet(
		t,
		findings,
		"jobq.go:3",
		"func New in public package jobq: entry functions live in container.go and only delegate; move it there or move its logic under internal/",
	)
	assertSnippet(
		t,
		findings,
		"jobq.go:7",
		"var cache in public package jobq: the API holds only exported types, constants and Err sentinels; move it under internal/",
	)
}

func TestPublicapiFunctionStructs(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "struct of four exported func fields in a public package is reported",
				files: map[string]string{"toolkit.go": "package arith\n\ntype Toolkit struct {\n\tSquare, Cube func(value int) (result int)\n\tDouble func(value int) (result int)\n\tHalve func(value int) (result int)\n}\n"},
				want:  []string{"R14 toolkit.go:3"},
			},
			{
				name:  "fields of a named func type of the package count",
				files: map[string]string{"op.go": "package arith\n\ntype Op func(value int) (result int)\n", "toolkit.go": "package arith\n\ntype Toolkit struct {\n\tSquare, Cube Op\n\tDouble Op\n\tHalve (Op)\n}\n"},
				want:  []string{"R14 toolkit.go:3"},
			},
			{
				name:  "options struct with a few callbacks passes",
				files: map[string]string{"clock.go": "package arith\n\nimport \"time\"\n\ntype Clock struct {\n\tNow func() (now time.Time)\n\tAfter func(delay time.Duration) (fired <-chan time.Time)\n\tLimit int\n}\n"},
				want:  nil,
			},
			{
				name:  "unexported func fields and other types pass",
				files: map[string]string{"toolkit.go": "package arith\n\ntype Toolkit struct {\n\tsquare, cube, double, halve func()\n}\n\ntype Op func()\n"},
				want:  nil,
			},
			{
				name:  "structs under internal are left to R28",
				files: map[string]string{"internal/kit/toolkit.go": "package kit\n\ntype Toolkit struct {\n\tA, B, C, D func()\n}\n"},
				want:  nil,
			},
		},
	)
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"toolkit.go": "package arith\n\ntype Toolkit struct {\n\tSquare, Cube, Double, Halve func()\n}\n"},
		),
		"toolkit.go:3",
		"type Toolkit in public package arith has 4 exported func fields (Square, Cube, Double, Halve): a struct of funcs is behavior, not API; declare an interface with these operations here, implement it with methods under internal/ and return it from container.go",
	)
}
