package siblingimports_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/siblingimports"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestSiblingimports(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "sibling business package import is reported at the import",
			files: map[string]string{
				"internal/a/service.go": "package a\n\nimport \"example.com/m/internal/b\"\n\nvar _ = b.X\n",
				"internal/b/handler.go": "package b\n\nfunc X() {}\n",
			},
			want: []string{"R16 internal/a/service.go:3"},
		},
		{
			name: "shared support package import passes",
			files: map[string]string{
				"internal/a/a.go":      "package a\n\nimport \"example.com/m/internal/errors\"\n\nvar _ = errors.ErrX\n",
				"internal/errors/e.go": "package errors\n\nimport \"errors\"\n\nvar ErrX = errors.New(\"x\")\n",
			},
			want: nil,
		},
		{
			name: "container.go wiring several siblings passes",
			files: map[string]string{
				"internal/app/container.go": wiring("app"),
				"internal/engine/e.go":      "package engine\n\nfunc X() {}\n",
				"internal/adapter/a.go":     "package adapter\n\nfunc Y() {}\n",
			},
			want: nil,
		},
		{
			name: "file named after its own package wiring several siblings passes",
			files: map[string]string{
				"internal/config/config.go": wiring("config"),
				"internal/engine/e.go":      "package engine\n\nfunc X() {}\n",
				"internal/adapter/a.go":     "package adapter\n\nfunc Y() {}\n",
			},
			want: nil,
		},
		{
			name: "unconventionally named wiring file gets one wiring finding",
			files: map[string]string{
				"internal/app/service.go": wiring("app"),
				"internal/engine/e.go":    "package engine\n\nfunc X() {}\n",
				"internal/adapter/a.go":   "package adapter\n\nfunc Y() {}\n",
			},
			want: []string{"R47 internal/app/service.go:1"},
		},
		{
			name: "package outside any internal tree is not checked",
			files: map[string]string{
				"cmd/a/a.go": "package a\n\nimport \"example.com/m/cmd/b\"\n\nvar _ = b.X\n",
				"cmd/b/b.go": "package b\n\nfunc X() {}\n",
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

func wiring(
	packageName string,
) (code string) {
	return "package " + packageName + "\n\nimport (\n\t\"example.com/m/internal/adapter\"\n\t\"example.com/m/internal/engine\"\n)\n\nvar _ = engine.X\n\nvar _ = adapter.Y\n"
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
	findings, err = siblingimports.Siblingimports(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Siblingimports() error = %v",
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

func TestSiblingimportsRejectsMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	_, err := siblingimports.Siblingimports(
		missing,
		nil,
	)
	if err == nil {
		t.Error("Siblingimports() error = nil, want one for a missing root")
	}
}
