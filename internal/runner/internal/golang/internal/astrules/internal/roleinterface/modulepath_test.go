package roleinterface_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/roleinterface"
)

func TestRoleinterfaceModulePath(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "module path is derived from the import that matches the longest directory",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"jobs/jobs.go":              "package jobs\n",
				"internal/run/run.go":       "package run\n\nimport \"example.com/m/internal/jobs\"\n\nvar _ jobs.Progress = nil\n",
			},
			want: nil,
		},
		{
			name: "external import whose path ends in a local directory is outvoted",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/run/run.go":       "package run\n\nimport \"example.com/m/internal/jobs\"\n\ntype Runner interface {\n\tjobs.Progress\n\tRun()\n}\n\nvar _ Runner = nil\n",
				"internal/run/other.go":     "package run\n\nimport \"example.com/m/internal/jobs\"\n\ntype Other interface {\n\tjobs.Progress\n\tStop()\n}\n\nvar _ Other = nil\n",
				"internal/run/vendor.go":    "package run\n\nimport jobs \"github.com/x/y/internal/jobs\"\n\nvar _ jobs.Progress = nil\n",
			},
			want: []string{onlyFound},
		},
		{
			name: "import of the module root resolves to the root package",
			files: map[string]string{
				"root.go":             "package m\n\ntype Job interface {\n\tProgress\n\tRun()\n}\n\ntype Progress interface {\n\tReport()\n}\n\nvar _ Job = nil\n",
				"internal/run/run.go": "package run\n\nimport (\n\t\"example.com/m\"\n\t\"example.com/m/internal/jobs\"\n)\n\nvar _ m.Progress = nil\n\nvar _ = jobs.X\n",
				"internal/jobs/x.go":  "package jobs\n\nvar X = 1\n",
			},
			want: nil,
		},
		{
			name: "files without local imports resolve only same-package uses",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/run/run.go":       "package run\n\nimport \"fmt\"\n\nvar _ = fmt.Sprint\n",
			},
			want: []string{onlyFound},
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
