package compositionroot_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/compositionroot"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	checkerInterface = "package checks\n\n// Checker runs a check.\ntype Checker interface {\n\tCheck() error\n}\n"
	goastPackage     = "package goast\n\nfunc A() {}\n"
	rustastPackage   = "package rustast\n\nfunc A() {}\n"
)

func TestCompositionroot(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "component without a composition root is reported at the directory",
			files: map[string]string{
				"internal/checks/goast/a.go":   goastPackage,
				"internal/checks/rustast/a.go": rustastPackage,
				"internal/checks/tsast/a.go":   "package tsast\n\nfunc A() {}\n",
			},
			want: []string{"R48 internal/checks"},
		},
		{
			name: "children nested under the component's own internal count as its packages",
			files: map[string]string{
				"internal/checks/internal/goast/a.go":   goastPackage,
				"internal/checks/internal/rustast/a.go": rustastPackage,
			},
			want: []string{"R48 internal/checks"},
		},
		{
			name: "composition root exposing an interface passes",
			files: map[string]string{
				"internal/checks/checks.go":    checkerInterface,
				"internal/checks/goast/a.go":   goastPackage,
				"internal/checks/rustast/a.go": rustastPackage,
			},
			want: nil,
		},
		{
			name: "composition root exposing a function named after the package passes",
			files: map[string]string{
				"internal/checks/checks.go":             "package checks\n\n// Checks runs every check.\nfunc Checks() {}\n",
				"internal/checks/internal/goast/a.go":   goastPackage,
				"internal/checks/internal/rustast/a.go": rustastPackage,
			},
			want: nil,
		},
		{
			name: "composition root exposing neither is reported at its package clause",
			files: map[string]string{
				"internal/checks/checks.go":    "package checks\n\n// Run runs every check.\nfunc Run() {}\n",
				"internal/checks/goast/a.go":   goastPackage,
				"internal/checks/rustast/a.go": rustastPackage,
			},
			want: []string{"R48 internal/checks/checks.go:1"},
		},
		{
			name: "composition root exporting a struct is reported at the type",
			files: map[string]string{
				"internal/checks/checks.go":    checkerInterface + "\n// Runner is a concrete runner.\ntype Runner struct{}\n",
				"internal/checks/goast/a.go":   goastPackage,
				"internal/checks/rustast/a.go": rustastPackage,
			},
			want: []string{"R13 internal/checks/checks.go:9"},
		},
		{
			name: "container.go is recognized as the composition root",
			files: map[string]string{
				"internal/app/container.go":          "package app\n\n// Container wires the app.\ntype Container interface {\n\tRun() error\n}\n",
				"internal/app/internal/engine/e.go":  "package engine\n\nfunc A() {}\n",
				"internal/app/internal/adapter/a.go": "package adapter\n\nfunc A() {}\n",
			},
			want: nil,
		},
		{
			name: "loose file beside the composition root is reported",
			files: map[string]string{
				"internal/checks/checks.go":     checkerInterface,
				"internal/checks/go_checker.go": "package checks\n\nfunc helper() {}\n",
				"internal/checks/goast/a.go":    goastPackage,
				"internal/checks/rustast/a.go":  rustastPackage,
			},
			want: []string{"R49 internal/checks/go_checker.go"},
		},
		{
			name: "test file beside the composition root passes",
			files: map[string]string{
				"internal/checks/checks.go":      checkerInterface,
				"internal/checks/checks_test.go": "package checks_test\n",
				"internal/checks/goast/a.go":     goastPackage,
				"internal/checks/rustast/a.go":   rustastPackage,
			},
			want: nil,
		},
		{
			name: "directory with a single child package is not a component",
			files: map[string]string{
				"internal/checks/goast/a.go": goastPackage,
			},
			want: nil,
		},
		{
			name: "literal internal directory is never a component",
			files: map[string]string{
				"internal/runner/r.go": "package runner\n\nfunc A() {}\n",
				"internal/engine/e.go": "package engine\n\nfunc A() {}\n",
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

	findings, err = compositionroot.Compositionroot(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Compositionroot() error = %v",
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

func TestCompositionrootRejectsMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	_, err := compositionroot.Compositionroot(
		missing,
		nil,
	)
	if err == nil {
		t.Error("Compositionroot() error = nil, want one for a missing root")
	}
}
