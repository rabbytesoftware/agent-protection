package compositionroot_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/compositionroot"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	modulePath       = "example.com/m"
	moduleRootImport = "."
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
			name: "module without container.go is reported at its directory",
			files: map[string]string{
				"internal/app/app.go":                  unit("app"),
				"internal/app/internal/engine/e.go":    unit("engine"),
				"internal/app/internal/adapter/ad.go":  unit("adapter"),
				"internal/app/internal/adapter/ad2.go": unit("adapter"),
			},
			want: []string{"R48 internal/app"},
		},
		{
			name: "one internal child is enough to make a module",
			files: map[string]string{
				"internal/app/app.go":               unit("app"),
				"internal/app/internal/engine/e.go": unit("engine"),
			},
			want: []string{"R48 internal/app"},
		},
		{
			name: "container.go wiring the internal children passes",
			files: map[string]string{
				"internal/app/container.go":           unit("app internal/app/internal/engine internal/app/internal/adapter"),
				"internal/app/app.go":                 unit("app"),
				"internal/app/internal/engine/e.go":   unit("engine"),
				"internal/app/internal/adapter/ad.go": unit("adapter"),
			},
			want: nil,
		},
		{
			name: "other module files importing its internal packages are reported at the import",
			files: map[string]string{
				"internal/app/container.go":              unit("app internal/app/internal/engine"),
				"internal/app/service.go":                unit("app fmt internal/app/internal/engine internal/app/internal/engine/sub"),
				"internal/app/internal/engine/e.go":      unit("engine"),
				"internal/app/internal/engine/sub/s.go":  unit("sub"),
				"internal/app/internal/engine/sub/s2.go": unit("sub"),
			},
			want: []string{"R48 internal/app/service.go:5", "R48 internal/app/service.go:6"},
		},
		{
			name: "module files may import packages outside their own internal tree",
			files: map[string]string{
				"internal/app/container.go":         unit("app internal/app/internal/engine"),
				"internal/app/service.go":           unit("app internal/other internal/app/models internal/apps/internal/x"),
				"internal/app/internal/engine/e.go": unit("engine"),
				"internal/app/models/m.go":          unit("models"),
				"internal/other/o.go":               unit("other"),
			},
			want: nil,
		},
		{
			name: "package main wires in main.go and is exempt",
			files: map[string]string{
				"cmd/tool/main.go":                unit("main cmd/tool/internal/flags"),
				"cmd/tool/run.go":                 unit("main cmd/tool/internal/flags"),
				"cmd/tool/internal/flags/f.go":    unit("flags"),
				"cmd/tool/internal/flags/more.go": unit("flags"),
			},
			want: nil,
		},
		{
			name: "directory without source is not a module",
			files: map[string]string{
				"internal/app/internal/engine/e.go":  unit("engine"),
				"internal/app/internal/engine/e2.go": unit("engine"),
			},
			want: nil,
		},
		{
			name: "plain child packages do not make a module",
			files: map[string]string{
				"internal/checks/checks.go":  unit("checks internal/checks/goast"),
				"internal/checks/goast/a.go": unit("goast"),
				"internal/checks/rust/a.go":  unit("rust"),
			},
			want: nil,
		},
		{
			name: "module root counts as a module",
			files: map[string]string{
				"m.go":                 unit("m internal/engine"),
				"internal/engine/e.go": unit("engine"),
			},
			want: []string{"R48 .", "R48 m.go:4"},
		},
		{
			name: "file named after the package is no longer a composition root",
			files: map[string]string{
				"internal/app/app.go":               unit("app internal/app/internal/engine"),
				"internal/app/internal/engine/e.go": unit("engine"),
			},
			want: []string{"R48 internal/app", "R48 internal/app/app.go:4"},
		},
		{
			name: "exported structs and files beside container.go are left to other rules",
			files: map[string]string{
				"internal/app/container.go":         "package app\n\ntype Runner struct{}\n\nfunc Run() {}\n",
				"internal/app/helper.go":            unit("app fmt"),
				"internal/app/internal/engine/e.go": unit("engine"),
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

func TestCompositionrootSnippets(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "nested module is named by its path",
			files: map[string]string{
				"internal/app/app.go":               unit("app internal/app/internal/engine"),
				"internal/app/internal/engine/e.go": unit("engine"),
			},
			want: []string{
				"module internal/app has internal packages but no container.go: add container.go and construct them there",
				"only container.go wires internal/app's internal packages",
			},
		},
		{
			name: "module root is named by its package",
			files: map[string]string{
				"shop.go":              unit("shop internal/engine"),
				"internal/engine/e.go": unit("engine"),
			},
			want: []string{
				"module shop has internal packages but no container.go: add container.go and construct them there",
				"only container.go wires shop's internal packages",
			},
		},
	}
	for _, tc := range cases {
		assertEqual(
			t,
			tc.name,
			snippets(checked(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}

func snippets(
	findings []report.Finding,
) (texts []string) {
	texts = []string{}
	for _, finding := range findings {
		texts = append(
			texts,
			finding.Snippet,
		)
	}

	return texts
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

func unit(
	spec string,
) (code string) {
	fields := strings.Fields(spec)
	code = "package " + fields[0] + "\n\nimport (\n"
	for _, imported := range fields[1:] {
		code += "\t\"" + importPath(imported) + "\"\n"
	}

	return code + ")\n\nfunc Run() {}\n"
}

func importPath(
	imported string,
) (path string) {
	if imported == moduleRootImport {
		return modulePath
	}

	if !strings.Contains(
		imported,
		"/",
	) {
		return imported
	}

	return modulePath + "/" + imported
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

	slices.Sort(got)
	assertEqual(
		t,
		name,
		got,
		want,
	)
}

func assertEqual(
	t *testing.T,
	name string,
	got []string,
	want []string,
) {
	t.Helper()
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: got %q, want %q",
			name,
			got,
			want,
		)
	}
}
