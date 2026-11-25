package interfacesize_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/interfacesize"
)

const (
	reader  = "package cube\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n\tSquare()\n}\n"
	writer  = "package cube\n\ntype Writer interface {\n\tStore()\n\tDelete()\n\tReset()\n}\n"
	squarer = "package cube\n\ntype Squarer interface {\n\tReader\n\tWriter\n}\n"
	bloated = "R28 internal/cube/squarer.go:3 interface Squarer has 6 methods once embedded Reader, Writer are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"
)

func TestInterfacesize(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "two embedded role interfaces adding up to six methods are reported",
			files: map[string]string{"internal/cube/reader.go": reader, "internal/cube/writer.go": writer, "internal/cube/squarer.go": squarer},
			want:  []string{bloated},
		},
		{
			name:  "embedding that stays within three methods passes",
			files: map[string]string{"internal/cube/squarer.go": "package cube\n\ntype Squarer interface {\n\tReader\n\tName()\n}\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n}\n"},
			want:  nil,
		},
		{
			name:  "interface with only direct methods over the cap is left to interfacebloat",
			files: map[string]string{"internal/cube/big.go": "package cube\n\ntype Big interface {\n\tA()\n\tB()\n\tC()\n\tD()\n}\n"},
			want:  nil,
		},
		{
			name:  "interface with more than three members including an embed is left to interfacebloat",
			files: map[string]string{"internal/cube/reader.go": reader, "internal/cube/big.go": "package cube\n\ntype Big interface {\n\tReader\n\tA()\n\tB()\n\tC()\n}\n"},
			want:  nil,
		},
		{
			name:  "unresolvable stdlib embeds count as zero methods",
			files: map[string]string{"internal/cube/stream.go": "package cube\n\nimport \"io\"\n\ntype Stream interface {\n\tio.Reader\n\tio.Writer\n\tio.Closer\n}\n"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			interfacesize.Interfacesize(parsedFiles(
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
			finding.Rule+" "+finding.Location+" "+finding.Snippet,
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
