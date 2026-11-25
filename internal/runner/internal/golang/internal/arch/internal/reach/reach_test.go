package reach_test

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/reach"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	modulePath       = "example.com/m"
	moduleRootImport = "."
)

func TestReach(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "outsider importing a package below a module is reported at the import",
			files: storeModule(map[string]string{
				"api/api.go": unit("api fmt store/models"),
			}),
			want: []string{"R15 api/api.go:5"},
		},
		{
			name: "outsider importing the module itself passes",
			files: storeModule(map[string]string{
				"api/api.go": unit("api store"),
			}),
			want: nil,
		},
		{
			name: "the module and its own packages may import below it",
			files: storeModule(map[string]string{
				"store/internal/memory/extra.go": unit("memory store/models"),
				"store/models/options/o.go":      unit("options store/models"),
			}),
			want: nil,
		},
		{
			name: "directory without internal children is not a module",
			files: map[string]string{
				"pkg/pkg.go":     unit("pkg"),
				"pkg/sub/sub.go": unit("sub"),
				"api/api.go":     unit("api pkg/sub"),
			},
			want: nil,
		},
		{
			name: "directory without source is not a module",
			files: map[string]string{
				"lib/internal/x/x.go": unit("x"),
				"lib/models/m.go":     unit("models"),
				"api/api.go":          unit("api lib/models"),
			},
			want: nil,
		},
		{
			name: "module root never counts as the reached module",
			files: map[string]string{
				"m.go":                 unit("m internal/engine"),
				"internal/engine/e.go": unit("engine"),
				"cmd/tool/main.go":     unit("main internal/engine"),
			},
			want: nil,
		},
		{
			name: "nested module is reached from inside its parent module",
			files: map[string]string{
				"app/container.go":                          unit("app app/internal/engine app/internal/web"),
				"app/internal/engine/container.go":          unit("engine app/internal/engine/internal/core"),
				"app/internal/engine/internal/core/core.go": unit("core"),
				"app/internal/engine/options/options.go":    unit("options"),
				"app/internal/web/web.go":                   unit("web app/internal/engine/options"),
			},
			want: []string{"R15 app/internal/web/web.go:4"},
		},
		{
			name: "leaf importing an uncle package passes",
			files: map[string]string{
				"golang/golang.go":                             unit("golang golang/internal/arch golang/internal/source"),
				"golang/internal/source/source.go":             unit("source"),
				"golang/internal/arch/arch.go":                 unit("arch golang/internal/arch/internal/roots"),
				"golang/internal/arch/internal/roots/roots.go": unit("roots golang/internal/source"),
			},
			want: nil,
		},
		{
			name: "production code importing an example is reported",
			files: map[string]string{
				"internal/app/app.go":              unit("app examples/demo internal/app/example/sample"),
				"examples/demo/demo.go":            unit("demo"),
				"internal/app/example/sample/s.go": unit("sample"),
			},
			want: []string{"R61 internal/app/app.go:4", "R61 internal/app/app.go:5"},
		},
		{
			name: "examples may import examples",
			files: map[string]string{
				"examples/demo/demo.go":   unit("demo examples/shared"),
				"examples/shared/s.go":    unit("shared"),
				"_examples/other/main.go": unit("main examples/shared"),
			},
			want: nil,
		},
		{
			name: "imports outside the module are ignored",
			files: storeModule(map[string]string{
				"api/api.go": "package api\n\nimport \"example.com/mx/store/models\"\n",
			}),
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

func TestReachSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		storeModule(map[string]string{
			"api/api.go":         unit("api store/models examples/demo"),
			"examples/demo/d.go": unit("demo"),
		}),
	)
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Snippet,
		)
	}

	assertEqual(
		t,
		"snippets",
		got,
		[]string{
			"reaches into module store (imports store/models): import store itself and use what its API exposes",
			"imports example package examples/demo: production code never depends on examples; move what it needs out of the example tree into a production package",
		},
	)
}

func TestReachRejectsMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	_, err := reach.Reach(
		missing,
		nil,
	)
	if err == nil {
		t.Error("Reach() error = nil, want one for a missing root")
	}
}

func storeModule(
	files map[string]string,
) (all map[string]string) {
	all = map[string]string{
		"store/container.go":         unit("store store/internal/memory"),
		"store/internal/memory/m.go": unit("memory store/models"),
		"store/models/models.go":     unit("models"),
	}
	maps.Copy(
		all,
		files,
	)
	return all
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

	findings, err = reach.Reach(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Reach() error = %v",
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
