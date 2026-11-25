package methodfile_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/methodfile"
)

const (
	cubeType    = "package cube\n\ntype cube struct{}\n"
	squareCode  = "package cube\n\nfunc (c *cube) Square() (r int) {\n\treturn 1\n}\n"
	cubeMethods = "package cube\n\ntype cube struct{}\n\nfunc (c *cube) Cube() (r int) {\n\treturn 1\n}\n"
	plainCode   = "package cube\n\nfunc helper() {}\n"
)

func TestMethodfile(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "methods split into numbered part files are reported",
			files: map[string]string{"cube/cube.go": cubeType, "cube/part1.go": squareCode, "cube/part2.go": squareCode},
			want: []string{
				"cube/part1.go",
				"cube/part1.go:3",
				"cube/part2.go",
				"cube/part2.go:3",
			},
		},
		{
			name:  "methods beside their type pass",
			files: map[string]string{"cube/cube.go": cubeMethods},
			want:  nil,
		},
		{
			name:  "methods in a file named after another word are reported once per receiver",
			files: map[string]string{"cube/math.go": "package cube\n\nfunc (c *cube) A() {}\n\nfunc (c *cube) B() {}\n"},
			want:  []string{"cube/math.go:3"},
		},
		{
			name:  "methods of two receivers in one file report the stranger",
			files: map[string]string{"cube/cube.go": "package cube\n\nfunc (c *cube) A() {}\n\nfunc (s *square) B() {}\n"},
			want:  []string{"cube/cube.go:5"},
		},
		{
			name:  "generic receiver matches its snake case file",
			files: map[string]string{"set/ordered_set.go": "package set\n\nfunc (s *OrderedSet[T]) Add(\n\tv T,\n) {\n}\n"},
			want:  nil,
		},
		{
			name:  "platform suffixed file holding its type's methods passes",
			files: map[string]string{"cube/cube_linux.go": squareCode, "cube/cube_windows_amd64.go": squareCode},
			want:  nil,
		},
		{
			name:  "methods in container.go pass",
			files: map[string]string{"cube/container.go": squareCode},
			want:  nil,
		},
		{
			name:  "methods in main.go of package main pass",
			files: map[string]string{"cmd/main.go": "package main\n\nfunc (a app) run() {}\n"},
			want:  nil,
		},
		{
			name:  "mock files are skipped",
			files: map[string]string{"cube/mocks/store1.go": squareCode},
			want:  nil,
		},
		{
			name:  "numbered segment is reported without methods",
			files: map[string]string{"cube/handler_2.go": plainCode},
			want:  []string{"cube/handler_2.go"},
		},
		{
			name:  "digit suffix of a sibling stem is reported",
			files: map[string]string{"cube/cube.go": plainCode, "cube/cube2.go": plainCode},
			want:  []string{"cube/cube2.go"},
		},
		{
			name: "digits that belong to the name pass",
			files: map[string]string{
				"hash/sha256.go":  plainCode,
				"hash/sha512.go":  plainCode,
				"codec/base64.go": plainCode,
				"api/v1.go":       plainCode,
				"net/ipv4.go":     plainCode,
				"sys/zsys_386.go": plainCode,
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			methodfile.Methodfile(parsedAll(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}

func TestMethodfileSnippets(
	t *testing.T,
) {
	findings := methodfile.Methodfile(parsedAll(
		t,
		map[string]string{"cube/part1.go": "package cube\n\nfunc (c *OrderItem) A() {}\n"},
	))
	want := []string{
		"file part1.go has a numbered name: name the file after the one type it holds (or the package's entry function) and merge or split by type, not by part",
		"methods of OrderItem live in part1.go: move them into order_item.go, the file that declares OrderItem",
	}
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Snippet,
		)
	}

	isMismatch := !slices.Equal(
		got,
		[]string{
			"R66 " + want[0],
			"R66 " + want[1],
		},
	)
	if isMismatch {
		t.Errorf(
			"snippets = %q, want %q",
			got,
			want,
		)
	}
}

func parsedAll(
	t *testing.T,
	sources map[string]string,
) (files []source.File) {
	t.Helper()
	for _, path := range slices.Sorted(maps.Keys(sources)) {
		files = append(
			files,
			parsed(
				t,
				path,
				sources[path],
			),
		)
	}

	return files
}

func parsed(
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
			"parse fixture: %v",
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

func assertLocations(
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
			finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: locations = %v, want %v",
			name,
			got,
			want,
		)
	}
}
