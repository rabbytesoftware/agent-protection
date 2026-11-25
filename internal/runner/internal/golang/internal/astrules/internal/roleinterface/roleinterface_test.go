package roleinterface_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/roleinterface"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	embedder  = "package jobs\n\ntype Job interface {\n\tProgress\n\tRun()\n}\n"
	progress  = "package jobs\n\ntype Progress interface {\n\tReport()\n}\n"
	onlyFound = "R60 internal/jobs/progress.go:3 interface Progress is only ever embedded: inline its methods into the interface that uses it"
)

func TestRoleinterface(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "exported interface only embedded in the same package is reported",
			files: map[string]string{"internal/jobs/job.go": embedder, "internal/jobs/progress.go": progress},
			want:  []string{onlyFound},
		},
		{
			name: "exported interface only embedded from another package is reported",
			files: map[string]string{
				"internal/jobs/progress.go": progress,
				"internal/queue/queue.go":   "package queue\n\nimport \"example.com/m/internal/jobs\"\n\ntype Job interface {\n\tjobs.Progress\n\tRun()\n}\n",
			},
			want: []string{onlyFound},
		},
		{
			name: "interface also used as a parameter type passes",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/jobs/run.go":      "package jobs\n\nfunc Track(\n\tp Progress,\n) {\n}\n",
			},
			want: nil,
		},
		{
			name: "interface used in a test file passes",
			files: map[string]string{
				"internal/jobs/job.go":           embedder,
				"internal/jobs/progress.go":      progress,
				"internal/jobs/progress_test.go": "package jobs\n\nvar _ Progress = nil\n",
			},
			want: nil,
		},
		{
			name:  "interface never referenced passes",
			files: map[string]string{"internal/jobs/progress.go": progress},
			want:  nil,
		},
		{
			name:  "unexported interface only embedded passes",
			files: map[string]string{"internal/jobs/job.go": "package jobs\n\ntype Job interface {\n\tprogress\n\tRun()\n}\n\ntype progress interface {\n\tReport()\n}\n"},
			want:  nil,
		},
		{
			name:  "interface that only re-embeds a qualified interface is reported",
			files: map[string]string{"internal/bus/bus.go": "package bus\n\nimport \"example.com/m/internal/pubsub\"\n\ntype Bus interface {\n\tpubsub.Bus\n}\n"},
			want:  []string{"R60 internal/bus/bus.go:5 interface Bus only re-embeds pubsub.Bus: use pubsub.Bus directly"},
		},
		{
			name:  "unexported interface that only re-embeds a local generic interface is reported",
			files: map[string]string{"internal/bus/bus.go": "package bus\n\ntype Sink[T any] interface {\n\tPut(\n\t\tv T,\n\t)\n}\n\ntype ints interface {\n\tSink[int]\n}\n\nvar _ ints = nil\n\nvar _ Sink[int] = nil\n"},
			want:  []string{"R60 internal/bus/bus.go:9 interface ints only re-embeds Sink[int]: use Sink[int] directly"},
		},
		{
			name:  "union constraint interface passes",
			files: map[string]string{"internal/num/num.go": "package num\n\ntype Number interface {\n\t~int | ~float64\n}\n\nfunc Sum[T Number]() {\n}\n"},
			want:  nil,
		},
		{
			name:  "interface with one embedded member and a method passes",
			files: map[string]string{"internal/bus/bus.go": "package bus\n\nimport \"io\"\n\ntype Stream interface {\n\tio.Reader\n\tName() string\n}\n\nvar _ Stream = nil\n"},
			want:  nil,
		},
		{
			name:  "alias of another interface passes",
			files: map[string]string{"internal/bus/bus.go": "package bus\n\nimport \"io\"\n\ntype Reader = io.Reader\n"},
			want:  nil,
		},
		{
			name:  "interface outside an internal tree passes",
			files: map[string]string{"pkg/bus/bus.go": "package bus\n\nimport \"io\"\n\ntype Reader interface {\n\tio.Reader\n}\n"},
			want:  nil,
		},
		{
			name:  "interface declared in a test file passes",
			files: map[string]string{"internal/bus/bus_test.go": "package bus\n\nimport \"io\"\n\ntype reader interface {\n\tio.Reader\n}\n"},
			want:  nil,
		},
		{
			name:  "interface declared in a mock passes",
			files: map[string]string{"internal/bus/mocks/bus.go": "package mocks\n\nimport \"io\"\n\ntype Reader interface {\n\tio.Reader\n}\n"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			roleinterface.Roleinterface(parsedFiles(
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

	return source.File{Path: path, Rel: path, FileSet: fileSet, Syntax: syntax}
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
