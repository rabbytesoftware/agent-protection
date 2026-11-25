package publicapi_test

import "testing"

func TestPublicapiReceiversAndNames(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "methods on generic receivers are named after the type",
				files: map[string]string{"jobq.go": "package jobq\n\ntype Box[T any] struct{}\n\ntype Pair[K, V any] struct{}\n\nfunc (b *Box[T]) String() string {\n\treturn \"\"\n}\n\nfunc (p (Pair[K, V])) Error() string {\n\treturn \"\"\n}\n\nfunc (b Box[T]) Get() {}\n"},
				want:  []string{"R14 jobq.go:15"},
			},
			{
				name:  "blank function passes",
				files: map[string]string{"jobq.go": "package jobq\n\nfunc _() {}\n"},
				want:  nil,
			},
			{
				name:  "every unexported name in a value spec is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nvar a, ErrB, c = 1, 2, 3\n"},
				want:  []string{"R14 jobq.go:3", "R14 jobq.go:3"},
			},
			{
				name:  "Err-prefixed unexported var is reported",
				files: map[string]string{"jobq.go": "package jobq\n\nvar errClosed = 1\n"},
				want:  []string{"R14 jobq.go:3"},
			},
		},
	)
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"jobq.go": "package jobq\n\ntype Box[T any] struct{}\n\nfunc (b Box[T]) Get() {}\n"},
		),
		"jobq.go:5",
		"method Box.Get in public package jobq: implementation lives under internal/; expose an interface here and construct the implementation in container.go",
	)
}
