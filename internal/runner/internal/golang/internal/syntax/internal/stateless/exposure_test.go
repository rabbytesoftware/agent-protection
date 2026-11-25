package stateless_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/stateless"
)

func TestStatelessRequiresInterfaceExposure(
	t *testing.T,
) {
	runner := "package p\n\ntype Runner interface {\n\tRun()\n}\n"
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "exported empty struct implementing an interface is reported",
			files: map[string]string{"p/runner.go": runner, "p/service.go": "package p\n\ntype Service struct{}\n\nfunc (Service) Run() {\n}\n"},
			want:  []string{"R11 p/service.go:3"},
		},
		{
			name:  "empty struct taken as a parameter is reported",
			files: map[string]string{"p/runner.go": runner, "p/service.go": "package p\n\ntype service struct{}\n\nfunc (service) Run() {\n}\n\nfunc start(\n\ts service,\n) {\n}\n"},
			want:  []string{"R11 p/service.go:3"},
		},
		{
			name:  "empty struct returned as the interface passes",
			files: map[string]string{"p/runner.go": runner, "p/service.go": "package p\n\ntype service struct{}\n\nfunc (service) Run() {\n}\n\nfunc New() (r Runner) {\n\treturn service{}\n}\n"},
			want:  nil,
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
