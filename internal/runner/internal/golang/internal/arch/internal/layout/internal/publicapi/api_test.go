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
