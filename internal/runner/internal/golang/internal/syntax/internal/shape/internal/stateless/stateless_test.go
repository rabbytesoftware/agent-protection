package stateless_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/stateless"
)

func TestStateless(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "empty struct with a method is reported at the type",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc (s service) Run() {\n}\n"},
			want:  []string{"R11 p/a.go:3"},
		},
		{
			name:  "empty mock with methods and a constructor passes",
			files: map[string]string{"p/mocks/bus.go": "package mocks\n\ntype Bus struct{}\n\nfunc New() *Bus {\n\treturn &Bus{}\n}\n\nfunc (b *Bus) Publish() {\n}\n"},
			want:  nil,
		},
		{
			name:  "method declared in a sibling file of the same package is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n", "p/b.go": "package p\n\nfunc (s *service) Run() {\n}\n"},
			want:  []string{"R11 p/a.go:3"},
		},
		{
			name:  "generic empty struct with a method is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype box[T any] struct{}\n\nfunc (b box[T]) Get() {\n}\n"},
			want:  []string{"R11 p/a.go:3"},
		},
		{
			name:  "method on a same-named type in another package passes",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n", "q/b.go": "package q\n\ntype service int\n\nfunc (s service) Run() {\n}\n"},
			want:  nil,
		},
		{
			name:  "struct with fields and methods passes",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct {\n\tname string\n}\n\nfunc (s service) Run() {\n}\n"},
			want:  nil,
		},
		{
			name:  "empty struct without methods passes",
			files: map[string]string{"p/a.go": "package p\n\ntype marker struct{}\n"},
			want:  nil,
		},
		{
			name:  "New returning an empty struct is reported at the constructor",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc New() *service {\n\treturn &service{}\n}\n"},
			want:  []string{"R12 p/a.go:5"},
		},
		{
			name:  "NewXxx returning an empty struct with an error is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc NewService() (s service, err error) {\n\treturn service{}, nil\n}\n"},
			want:  []string{"R12 p/a.go:5"},
		},
		{
			name:  "constructor that can also return nil passes",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc New(\n\tname string,\n) *service {\n\tif name == \"\" {\n\t\treturn nil\n\t}\n\treturn &service{}\n}\n"},
			want:  nil,
		},
		{
			name:  "constructor building a struct with state passes",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct {\n\tname string\n}\n\nfunc New() *service {\n\treturn &service{name: \"x\"}\n}\n"},
			want:  nil,
		},
		{
			name:  "function returning an empty struct is reported whatever its name",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc MakeService() service {\n\treturn service{}\n}\n\nfunc Default() *service {\n\treturn &service{}\n}\n"},
			want:  []string{"R12 p/a.go:5", "R12 p/a.go:9"},
		},
		{
			name:  "struct with only blank fields and a method is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype Pricer struct {\n\t_ struct{}\n\t_, _ int\n}\n\nfunc (Pricer) Price() {\n}\n"},
			want:  []string{"R11 p/a.go:3"},
		},
		{
			name:  "function returning a struct with only blank fields is reported",
			files: map[string]string{"p/a.go": "package p\n\ntype Pricer struct {\n\t_ struct{}\n}\n\nfunc MakePricer() (pricer Pricer) {\n\treturn Pricer{}\n}\n"},
			want:  []string{"R12 p/a.go:7"},
		},
		{
			name:  "struct mixing a blank and a named field passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Pricer struct {\n\t_    struct{}\n\trate int\n}\n\nfunc (Pricer) Price() {\n}\n\nfunc MakePricer() (pricer Pricer) {\n\treturn Pricer{}\n}\n"},
			want:  nil,
		},
		{
			name:  "struct with an embedded field passes",
			files: map[string]string{"p/a.go": "package p\n\ntype Pricer struct {\n\tbase\n}\n\nfunc (Pricer) Price() {\n}\n"},
			want:  nil,
		},
		{
			name:  "function with a bare return passes",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc run() {\n\treturn\n}\n"},
			want:  nil,
		},
		{
			name:  "generic constructors and a parenthesized receiver of empty structs are reported",
			files: map[string]string{"p/a.go": "package p\n\ntype box[T any] struct{}\n\ntype pair[K comparable, V any] struct{}\n\nfunc MakeBox() box[int] {\n\treturn box[int]{}\n}\n\nfunc MakePair() pair[string, int] {\n\treturn pair[string, int]{}\n}\n\nfunc (b (box[T])) Get() {\n}\n"},
			want:  []string{"R11 p/a.go:3", "R12 p/a.go:7", "R12 p/a.go:11"},
		},
		{
			name:  "method returning an empty struct is not a constructor",
			files: map[string]string{"p/a.go": "package p\n\ntype marker struct{}\n\ntype holder struct {\n\tname string\n}\n\nfunc (h holder) Marker() marker {\n\treturn marker{}\n}\n"},
			want:  nil,
		},
		{
			name:  "empty struct with both a method and a constructor is reported twice",
			files: map[string]string{"p/a.go": "package p\n\ntype service struct{}\n\nfunc New() service {\n\treturn service{}\n}\n\nfunc (s service) Run() {\n}\n"},
			want:  []string{"R11 p/a.go:3", "R12 p/a.go:5"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stateless.Stateless(parsedFiles(
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
