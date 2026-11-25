package dependency_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDependency(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		fails bool
		want  []string
	}{
		{
			name: "every child check contributes its findings in check order",
			files: map[string]string{
				"go.mod":                  "module example.com/m\n\ngo 1.24\n",
				"internal/a/service.go":   "package a\n\nimport \"example.com/m/internal/b\"\n\nfunc A() {\n\tb.B()\n}\n",
				"internal/a/misc_test.go": "package a_test\n",
				"internal/b/b.go":         "package b\n\nimport \"example.com/m/internal/c\"\n\nfunc B() {\n\tc.C()\n}\n",
				"internal/b/b_test.go":    "package b_test\n\nfunc TestB() {}\n",
				"internal/c/c.go":         "package c\n\nfunc C() {}\n",
				"internal/c/c_test.go":    "package c_test\n\nfunc TestC() {}\n",
			},
			want: []string{
				"R16 internal/a/service.go:3",
				"R19 internal/b",
				"R19 internal/c",
			},
		},
		{
			name: "module without go.mod fails instead of reporting partial findings",
			files: map[string]string{
				"internal/a/a.go": "package a\n",
			},
			fails: true,
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertRules(
			t,
			tc.name,
			tc.files,
			tc.fails,
			tc.want,
		)
	}
}

func assertRules(
	t *testing.T,
	name string,
	files map[string]string,
	fails bool,
	want []string,
) {
	t.Helper()
	root := module(
		t,
		files,
	)
	parsed, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"%s: parse fixture module: %v",
			name,
			err,
		)
	}

	findings, err := dependency.Dependency(
		root,
		parsed,
	)
	hasFailed := err != nil
	if hasFailed != fails {
		t.Fatalf(
			"%s: Dependency() error = %v, want failure %v",
			name,
			err,
			fails,
		)
	}

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

func module(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
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
