package encapsulation_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/encapsulation"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestEncapsulation(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "package visible to the whole module with one importer is reported",
			files: map[string]string{
				"internal/checks/pipeline.go": "package checks\n\nimport \"example.com/m/internal/goast\"\n\nvar _ = goast.X\n",
				"internal/goast/g.go":         "package goast\n\nfunc X() {}\n",
			},
			want: []string{"R19 internal/goast"},
		},
		{
			name: "package nested under its importer's internal passes",
			files: map[string]string{
				"internal/checks/pipeline.go":         "package checks\n\nimport \"example.com/m/internal/checks/internal/goast\"\n\nvar _ = goast.X\n",
				"internal/checks/internal/goast/g.go": "package goast\n\nfunc X() {}\n",
			},
			want: nil,
		},
		{
			name: "package with two importers passes",
			files: map[string]string{
				"internal/a/a.go":      "package a\n\nimport \"example.com/m/internal/shared\"\n\nvar _ = shared.X\n",
				"internal/b/b.go":      "package b\n\nimport \"example.com/m/internal/shared\"\n\nvar _ = shared.X\n",
				"internal/shared/s.go": "package shared\n\nfunc X() {}\n",
			},
			want: nil,
		},
		{
			name: "package imported only by package main passes",
			files: map[string]string{
				"cmd/app/main.go":      "package main\n\nimport \"example.com/m/internal/runner\"\n\nfunc main() {\n\trunner.Run()\n}\n",
				"internal/runner/r.go": "package runner\n\nfunc Run() {}\n",
			},
			want: nil,
		},
		{
			name: "shared support package with one importer passes",
			files: map[string]string{
				"internal/a/a.go":      "package a\n\nimport \"example.com/m/internal/report\"\n\nvar _ = report.X\n",
				"internal/report/r.go": "package report\n\nfunc X() {}\n",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				tc.files,
			),
			tc.want,
		)
	}
}

func checked(
	t *testing.T,
	files map[string]string,
) (findings []report.Finding) {
	t.Helper()
	root := module(
		t,
		files,
	)
	parsed, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"parse fixture module: %v",
			err,
		)
	}
	findings, err = encapsulation.Encapsulation(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Encapsulation() error = %v",
			err,
		)
	}
	return findings
}

func module(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	write(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"module example.com/m\n\ngo 1.24\n",
	)
	for name, content := range files {
		write(
			t,
			filepath.Join(
				root,
				name,
			),
			content,
		)
	}
	return root
}

func write(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"mkdir %s: %v",
			path,
			err,
		)
	}
	err = os.WriteFile(
		path,
		[]byte(content),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
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
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func TestEncapsulationRejectsMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	_, err := encapsulation.Encapsulation(
		missing,
		nil,
	)
	if err == nil {
		t.Error("Encapsulation() error = nil, want one for a missing root")
	}
}
