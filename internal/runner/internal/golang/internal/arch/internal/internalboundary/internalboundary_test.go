package internalboundary_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/internalboundary"
)

const rulesPackage = "package rules\n\nfunc Rules() {}\n"

func TestInternalboundary(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "file with behavior loose in internal is reported",
			files: map[string]string{
				"internal/lint/internal/golangci.go": "package lint\n\nfunc Run() {}\n",
				"internal/rules/rules.go":            rulesPackage,
			},
			want: []string{"R50 internal/lint/internal/golangci.go"},
		},
		{
			name: "every loose file with behavior is reported",
			files: map[string]string{
				"internal/checks/internal/go.go":   "package pipeline\n\nfunc CheckGo() {}\n",
				"internal/checks/internal/rust.go": "package pipeline\n\nfunc CheckRust() {}\n",
				"internal/rules/rules.go":          rulesPackage,
			},
			want: []string{"R50 internal/checks/internal/go.go", "R50 internal/checks/internal/rust.go"},
		},
		{
			name: "declaration-only file loose in internal passes",
			files: map[string]string{
				"internal/golang/internal/component.go": "package component\n\ntype Analyzer interface {\n\tRun() error\n}\n",
				"internal/rules/rules.go":               rulesPackage,
			},
			want: nil,
		},
		{
			name: "sole child of an internal directory is reported",
			files: map[string]string{
				"internal/runner/internal/checks/checks.go": "package checks\n\nfunc Checks() {}\n",
				"internal/rules/rules.go":                   rulesPackage,
			},
			want: []string{"R51 internal/runner/internal/checks"},
		},
		{
			name: "child whose source sits deeper still counts as one child",
			files: map[string]string{
				"internal/runner/internal/checks/internal/a/a.go": "package a\n\nfunc A() {}\n",
				"internal/runner/internal/checks/internal/b/b.go": "package b\n\nfunc B() {}\n",
				"internal/rules/rules.go":                         rulesPackage,
			},
			want: []string{"R51 internal/runner/internal/checks"},
		},
		{
			name: "several sibling children pass",
			files: map[string]string{
				"internal/runner/internal/golang/golang.go": "package golang\n\nfunc Golang() {}\n",
				"internal/runner/internal/rust/rust.go":     "package rust\n\nfunc Rust() {}\n",
				"internal/rules/rules.go":                   rulesPackage,
			},
			want: nil,
		},
		{
			name: "module without internal directories passes",
			files: map[string]string{
				"golang.go": "package example\n\nfunc Run() {}\n",
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
	findings, err := internalboundary.Internalboundary(module(
		t,
		files,
	))
	if err != nil {
		t.Fatalf(
			"Internalboundary() error = %v",
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
