package concretetype_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/concretetype"
)

func TestConcretetype(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "exported struct with a pointer-receiver method is reported at the type",
			files: map[string]string{"p/a.go": "package p\n\ntype Store struct {\n\tn int\n}\n\nfunc (s *Store) Put() {\n}\n"},
			want:  []string{"R13 p/a.go:3"},
		},
		{
			name:  "pointer-receiver method declared in a sibling file is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype Store struct {\n\tn int\n}\n", "p/b.go": "package p\n\nfunc (s *Store) Put() {\n}\n"},
			want:  []string{"R13 p/a.go:3"},
		},
		{
			name:  "generic exported struct with a pointer-receiver method is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype Cache[K comparable, V any] struct {\n\tm map[K]V\n}\n\nfunc (c *Cache[K, V]) Get() {\n}\n"},
			want:  []string{"R13 p/a.go:3"},
		},
		{
			name:  "single type parameter struct with a parenthesized pointer receiver is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype Box[T any] struct {\n\tv T\n}\n\nfunc (b (*Box[T])) Set() {\n}\n"},
			want:  []string{"R13 p/a.go:3"},
		},
		{
			name:  "unexported struct with pointer-receiver methods passes",
			files: map[string]string{"p/a.go": "package p\n\ntype store struct {\n\tn int\n}\n\nfunc (s *store) Put() {\n}\n"},
			want:  nil,
		},
		{
			name:  "exported value type with only value-receiver methods passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Point struct {\n\tX int\n}\n\nfunc (p Point) Add() {\n}\n"},
			want:  nil,
		},
		{
			name:  "exported struct without methods passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Config struct {\n\tName string\n}\n"},
			want:  nil,
		},
		{
			name:  "exported error type with a pointer-receiver Error method passes",
			files: map[string]string{"p/a.go": "package p\n\ntype NotFoundError struct {\n\tKey string\n}\n\nfunc (e *NotFoundError) Error() string {\n\treturn e.Key\n}\n\nfunc (e *NotFoundError) Unwrap() error {\n\treturn nil\n}\n"},
			want:  nil,
		},
		{
			name:  "exported non-struct type with a pointer-receiver method passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Counter int\n\nfunc (c *Counter) Inc() {\n}\n"},
			want:  nil,
		},
		{
			name:  "same-named pointer receiver in another package passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Store struct {\n\tn int\n}\n", "q/b.go": "package q\n\ntype Store struct {\n\tn int\n}\n\nfunc (s Store) Put() {\n}\n\ntype impl struct{}\n\nfunc (i *impl) Put() {\n}\n"},
			want:  nil,
		},
		{
			name:  "package main passes",
			files: map[string]string{"cmd/a/main.go": "package main\n\ntype Server struct {\n\tn int\n}\n\nfunc (s *Server) Run() {\n}\n"},
			want:  nil,
		},
		{
			name:  "mock package passes",
			files: map[string]string{"p/mocks/bus.go": "package mocks\n\ntype Bus struct {\n\tn int\n}\n\nfunc (b *Bus) Publish() {\n}\n"},
			want:  nil,
		},
		{
			name:  "test file passes",
			files: map[string]string{"p/a_test.go": "package p\n\ntype Recorder struct {\n\tn int\n}\n\nfunc (r *Recorder) Record() {\n}\n"},
			want:  nil,
		},
		{
			name:  "type alias to a struct passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Store = struct{ n int }\n\ntype impl struct{}\n\nfunc (i *impl) Put() {\n}\n"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			concretetype.Concretetype(parsedFiles(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}

func parsedFiles(
	t *testing.T,
	files map[string]string,
) (parsed []source.File) {
	t.Helper()
	for _, path := range slices.Sorted(maps.Keys(files)) {
		parsed = append(
			parsed,
			parsedFile(
				t,
				path,
				files[path],
			),
		)
	}

	return parsed
}

func parsedFile(
	t *testing.T,
	path string,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		path,
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture %s: %v",
			path,
			err,
		)
	}

	return source.File{
		Path:    path,
		Rel:     path,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}

func assertFindings(
	t *testing.T,
	name string,
	findings []report.Finding,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
