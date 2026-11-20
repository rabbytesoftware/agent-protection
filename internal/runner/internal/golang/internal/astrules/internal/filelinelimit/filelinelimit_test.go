package filelinelimit_test

import (
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/filelinelimit"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
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
				code,
			)}),
			tc.want,
		)
	}
}

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		"p.go",
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}
	return source.File{Path: "p.go", Rel: "p.go", FileSet: fileSet, Syntax: syntax}
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
