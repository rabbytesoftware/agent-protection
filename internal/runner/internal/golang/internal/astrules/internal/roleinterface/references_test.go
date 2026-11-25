package roleinterface_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/roleinterface"
)

func TestRoleinterfaceReferences(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "use through an explicit import alias passes",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/run/run.go":       "package run\n\nimport j \"example.com/m/internal/jobs\"\n\nvar _ j.Progress = nil\n",
			},
			want: nil,
		},
		{
			name: "use through the declared package name of a differently named directory passes",
			files: map[string]string{
				"internal/v2/job.go":      embedder,
				"internal/v2/progress.go": progress,
				"internal/run/run.go":     "package run\n\nimport \"example.com/m/internal/v2\"\n\nvar _ jobs.Progress = nil\n",
			},
			want: nil,
		},
		{
			name: "use from an external test package passes",
			files: map[string]string{
				"internal/jobs/job.go":       embedder,
				"internal/jobs/progress.go":  progress,
				"internal/jobs/jobs_test.go": "package jobs_test\n\nimport \"example.com/m/internal/jobs\"\n\nvar _ jobs.Progress = nil\n",
			},
			want: nil,
		},
		{
			name: "bare name in an external test package does not count as a use",
			files: map[string]string{
				"internal/jobs/job.go":       embedder,
				"internal/jobs/progress.go":  progress,
				"internal/jobs/jobs_test.go": "package jobs_test\n\ntype Progress interface{}\n\nvar _ Progress = nil\n",
			},
			want: []string{onlyFound},
		},
		{
			name: "field selector with the same name does not count as a use",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/jobs/state.go":    "package jobs\n\ntype state struct {\n\tProgress int\n}\n\nfunc read(\n\ts state,\n) (n int) {\n\treturn s.Progress\n}\n",
			},
			want: []string{onlyFound},
		},
		{
			name: "same-named interface used in another package does not count as a use",
			files: map[string]string{
				"internal/jobs/job.go":       embedder,
				"internal/jobs/progress.go":  progress,
				"internal/other/progress.go": "package other\n\ntype Progress interface {\n\tReport()\n}\n\nvar _ Progress = nil\n",
			},
			want: []string{onlyFound},
		},
		{
			name: "blank and dot imports are ignored",
			files: map[string]string{
				"internal/jobs/job.go":      embedder,
				"internal/jobs/progress.go": progress,
				"internal/run/run.go":       "package run\n\nimport (\n\t_ \"example.com/m/internal/jobs\"\n\t. \"example.com/m/internal/jobs\"\n)\n\nvar _ Job = nil\n",
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
