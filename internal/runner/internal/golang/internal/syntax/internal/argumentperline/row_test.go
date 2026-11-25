package argumentperline_test

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/argumentperline"
)

const testFileName = "p_test.go"

func TestTestRows(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "inline rows of a slice of structs pass as whole rows",
			code: "package p\n\nvar cases = []struct {\n\tname string\n\twant []string\n}{\n\t{name: \"a\", want: []string{\"x\", \"y\", \"z\", \"w\"}},\n\t{name: \"b\", want: nil},\n}\n",
			want: nil,
		},
		{
			name: "inline pointer rows and map rows pass",
			code: "package p\n\nvar a = []*T{\n\t&T{X: g(), Y: h()},\n}\nvar b = map[string]T{\n\t\"one\": {X: g(), Y: h()},\n}\n",
			want: nil,
		},
		{
			name: "a row spread over several lines is held to one element per line",
			code: "package p\n\nvar a = []T{\n\t{X: g(),\n\t\tY: h()},\n}\n",
			want: []string{"R52 p_test.go:4"},
		},
		{
			name: "the table itself still needs one row per line",
			code: "package p\n\nvar a = []T{{X: 1}, {X: 2}}\n",
			want: []string{"R52 p_test.go:3"},
		},
		{
			name: "inline struct literal outside a table is reported in tests",
			code: "package p\n\nvar a = T{X: g(), Y: h()}\n",
			want: []string{"R52 p_test.go:3"},
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			argumentperline.Argumentperline([]source.File{parsedTest(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsedTest(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		testFileName,
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
		Path:    testFileName,
		Rel:     testFileName,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}

func TestEditsKeepTestRows(
	t *testing.T,
) {
	code := "package p\n\nvar cases = []T{\n\t{name: \"a\", want: []string{\"x\", \"y\", \"z\", \"w\"}},\n}\n"
	file, err := source.Parse(
		"",
		testFileName,
		[]byte(code),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	edits := argumentperline.Edits(file)
	if len(edits) > 0 {
		t.Errorf(
			"Edits() = %v, want inline test rows left alone",
			edits,
		)
	}
}
