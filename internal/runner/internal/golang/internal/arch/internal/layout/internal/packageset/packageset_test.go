package packageset_test

import (
	"go/ast"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/packageset"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestPackageset(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "sibling whose name extends another sibling is reported",
			files: map[string]string{"internal/dispatch/d.go": "package dispatch\n", "internal/dispatcher/d.go": "package dispatcher\n"},
			want:  []string{"R58 internal/dispatcher dispatcher and its sibling dispatch split one concept: merge them or name each for what it provides"},
		},
		{
			name:  "longer sibling names its closest prefix sibling",
			files: map[string]string{"x/mock/m.go": "package mock\n", "x/mockbus/m.go": "package mockbus\n", "x/mockbuses/m.go": "package mockbuses\n"},
			want: []string{
				"R58 x/mockbus mockbus and its sibling mock split one concept: merge them or name each for what it provides",
				"R58 x/mockbuses mockbuses and its sibling mockbus split one concept: merge them or name each for what it provides",
			},
		},
		{
			name:  "child package whose name extends its parent passes",
			files: map[string]string{"store/s.go": "package store\n", "store/storetest/t.go": "package storetest\n"},
			want:  nil,
		},
		{
			name:  "prefix-named directories under different parents pass",
			files: map[string]string{"a/dispatch/d.go": "package dispatch\n", "b/dispatcher/d.go": "package dispatcher\n"},
			want:  nil,
		},
		{
			name:  "prefix sibling directory without go files passes",
			files: map[string]string{"internal/dispatch/README.md": "x\n", "internal/dispatcher/d.go": "package dispatcher\n"},
			want:  nil,
		},
		{
			name: "duplicate package names are reported after the first by path",
			files: map[string]string{
				"internal/eventstore/models/m.go": "package models\n",
				"internal/a/models/m.go":          "package models\n",
				"pkg/models/m.go":                 "package models\n",
			},
			want: []string{
				"R59 internal/eventstore/models package name models is also used by internal/a/models: give each package a name that says what it provides",
				"R59 pkg/models package name models is also used by internal/a/models: give each package a name that says what it provides",
			},
		},
		{
			name:  "duplicate declared names in differently named directories are reported",
			files: map[string]string{"a/v1/x.go": "package store\n", "b/v2/x.go": "package store\n"},
			want:  []string{"R59 b/v2 package name store is also used by a/v1: give each package a name that says what it provides"},
		},
		{
			name:  "several main packages pass",
			files: map[string]string{"cmd/a/main.go": "package main\n", "cmd/b/main.go": "package main\n"},
			want:  nil,
		},
		{
			name:  "external test package beside a same-named package passes",
			files: map[string]string{"a/store/s.go": "package store\n", "a/store/s_test.go": "package store_test\n", "b/x/x_test.go": "package store\n"},
			want:  nil,
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

func TestPackagesetRejectsUnrelatablePaths(
	t *testing.T,
) {
	file := source.File{
		Path: filepath.Join(
			t.TempDir(),
			"a",
			"a.go",
		),
		Syntax: &ast.File{Name: ast.NewIdent("a")},
	}
	_, err := packageset.Packageset(
		"relative",
		[]source.File{file},
	)
	if err == nil {
		t.Error("Packageset() error = nil, want one for a path outside the root")
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

	tests, err := source.Tests(root)
	if err != nil {
		t.Fatalf(
			"parse fixture tests: %v",
			err,
		)
	}

	findings, err = packageset.Packageset(
		root,
		slices.Concat(
			parsed,
			tests,
		),
	)
	if err != nil {
		t.Fatalf(
			"Packageset() error = %v",
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
			finding.Rule+" "+finding.Location+" "+finding.Snippet,
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
