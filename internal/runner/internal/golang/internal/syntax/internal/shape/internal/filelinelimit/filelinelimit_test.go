package filelinelimit_test

import (
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/filelinelimit"
)

const packageClause = "package p\n"

func TestFilelinelimit(
	t *testing.T,
) {
	cases := []struct {
		name  string
		lines int
		want  []string
	}{
		{
			name:  "file one line over the limit is reported at the package clause",
			lines: 501,
			want:  []string{"p.go:1"},
		},
		{
			name:  "file exactly at the limit passes",
			lines: 500,
			want:  nil,
		},
		{
			name:  "short file passes",
			lines: 1,
			want:  nil,
		},
	}
	for _, tc := range cases {
		code := packageClause + strings.Repeat(
			"// line\n",
			tc.lines-1,
		)
		assertLocations(
			t,
			tc.name,
			filelinelimit.Filelinelimit([]source.File{parsed(
				t,
				"p.go",
				code,
			)}),
			tc.want,
		)
	}
}

func TestFilelinelimitSnippets(
	t *testing.T,
) {
	code := packageClause + strings.Repeat(
		"// line\n",
		500,
	)
	findings := filelinelimit.Filelinelimit([]source.File{
		parsed(
			t,
			"p.go",
			code,
		),
		parsed(
			t,
			"p_test.go",
			code,
		),
	})
	want := []string{
		"501 lines, over the 500 line limit: split the file by responsibility",
		"501 lines, over the 500 line limit, test files included: a test file this long means its source file holds more than one responsibility; split the source by responsibility and move each group of tests into the x_test.go of the file it exercises (benchmarks may go to x_bench_test.go)",
	}
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Snippet,
		)
	}

	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"snippets = %q, want %q",
			got,
			want,
		)
	}
}

func parsed(
	t *testing.T,
	rel string,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		rel,
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
		Path:    rel,
		Rel:     rel,
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
