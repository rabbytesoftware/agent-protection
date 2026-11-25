package siblingimport_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/siblingimport"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	modulePath       = "example.com/m"
	moduleRootImport = "."
	declarationOnly  = "package contracts\n\ntype Order struct{}\n\ntype Store interface {\n\tSave() error\n}\n"
	errorsOnly       = "package failures\n\ntype NotFound struct{}\n\nfunc (NotFound) Error() (message string) {\n\treturn \"not found\"\n}\n"
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
			name: "leaf importing one foundation sibling passes",
			files: map[string]string{
				"internal/a/a.go": unit("a internal/b"),
				"internal/b/b.go": unit("b"),
			},
			want: nil,
		},
		{
			name: "importing a sibling that imports a sibling is reported at the import",
			files: map[string]string{
				"internal/a/a.go": unit("a internal/b"),
				"internal/b/b.go": unit("b internal/c"),
				"internal/c/c.go": unit("c"),
			},
			want: []string{"R16 internal/a/a.go:4"},
		},
		{
			name: "chains are reported in public trees too",
			files: map[string]string{
				"cmd/a/a.go": unit("a cmd/b"),
				"cmd/b/b.go": unit("b cmd/c"),
				"cmd/c/c.go": unit("c"),
			},
			want: []string{"R16 cmd/a/a.go:4"},
		},
		{
			name: "container.go gets no exemption",
			files: map[string]string{
				"internal/app/container.go": unit("app internal/b"),
				"internal/b/b.go":           unit("b internal/c"),
				"internal/c/c.go":           unit("c"),
			},
			want: []string{"R16 internal/app/container.go:4"},
		},
		{
			name: "shared support names get no exemption",
			files: map[string]string{
				"internal/a/a.go":           unit("a internal/errors"),
				"internal/errors/errors.go": unit("errors internal/c"),
				"internal/c/c.go":           unit("c"),
			},
			want: []string{"R16 internal/a/a.go:4"},
		},
		{
			name: "package importing two siblings is reported once at its directory",
			files: map[string]string{
				"internal/app/app.go":     unit("app internal/engine"),
				"internal/app/run.go":     unit("app internal/adapter internal/engine"),
				"internal/engine/e.go":    unit("engine"),
				"internal/adapter/ad.go":  unit("adapter"),
				"internal/adapter/ad2.go": unit("adapter"),
			},
			want: []string{"R47 internal/app"},
		},
		{
			name: "wiring and chains are both reported",
			files: map[string]string{
				"internal/app/app.go":    unit("app internal/adapter internal/engine"),
				"internal/engine/e.go":   unit("engine"),
				"internal/adapter/ad.go": unit("adapter internal/engine"),
			},
			want: []string{"R16 internal/app/app.go:4", "R47 internal/app"},
		},
		{
			name: "declaration-only siblings do not count towards wiring",
			files: map[string]string{
				"internal/app/app.go":            unit("app internal/contracts internal/failures internal/engine"),
				"internal/engine/e.go":           unit("engine"),
				"internal/contracts/contract.go": declarationOnly,
				"internal/failures/failures.go":  errorsOnly,
			},
			want: nil,
		},
		{
			name: "importer below a sibling is lifted to that sibling",
			files: map[string]string{
				"internal/alpha/alpha.go":               unit("alpha"),
				"internal/alpha/internal/deep/scan.go":  unit("deep internal/beta"),
				"internal/beta/beta.go":                 unit("beta internal/gamma"),
				"internal/gamma/gamma.go":               unit("gamma"),
				"internal/alpha/internal/store/s.go":    unit("store"),
				"internal/alpha/internal/deep/other.go": unit("deep internal/alpha/internal/store"),
			},
			want: []string{"R16 internal/alpha/internal/deep/scan.go:4"},
		},
		{
			name: "imported sibling's own subtree is lifted when deciding it is a foundation",
			files: map[string]string{
				"internal/alpha/alpha.go":         unit("alpha internal/beta"),
				"internal/beta/beta.go":           unit("beta"),
				"internal/beta/refund/refund.go":  unit("refund internal/gamma"),
				"internal/gamma/gamma.go":         unit("gamma"),
				"internal/beta/refund/refund2.go": unit("refund"),
			},
			want: []string{"R16 internal/alpha/alpha.go:4"},
		},
		{
			name: "lifted imports do not count towards wiring",
			files: map[string]string{
				"internal/alpha/alpha.go":     unit("alpha internal/beta"),
				"internal/alpha/sub/sub.go":   unit("sub internal/gamma"),
				"internal/beta/beta.go":       unit("beta"),
				"internal/gamma/gamma.go":     unit("gamma"),
				"internal/alpha/sub/other.go": unit("sub"),
			},
			want: nil,
		},
		{
			name: "leaf importing an uncle foundation passes",
			files: map[string]string{
				"golang/golang.go":                             unit("golang golang/internal/arch golang/internal/source"),
				"golang/internal/source/source.go":             unit("source"),
				"golang/internal/arch/arch.go":                 unit("arch golang/internal/arch/internal/roots golang/internal/source"),
				"golang/internal/arch/internal/roots/roots.go": unit("roots golang/internal/source"),
				"golang/internal/arch/internal/pairs/pairs.go": unit("pairs golang/internal/source"),
			},
			want: nil,
		},
		{
			name: "importing the module root or a child is not a sibling import",
			files: map[string]string{
				"m.go":                 unit("m internal/a internal/b"),
				"cmd/tool/main.go":     unit("main . cmd/tool/flags"),
				"cmd/tool/flags/f.go":  unit("flags"),
				"internal/a/a.go":      unit("a"),
				"internal/b/b.go":      unit("b"),
				"internal/b/sub/s.go":  unit("sub internal/b"),
				"internal/b/sub/s2.go": unit("sub"),
			},
			want: nil,
		},
		{
			name: "imports outside the module are ignored",
			files: map[string]string{
				"internal/a/a.go": "package a\n\nimport (\n\t\"fmt\"\n\t\"example.com/mx/internal/b\"\n\t\"example.com/mx/internal/c\"\n)\n",
				"internal/b/b.go": unit("b internal/c"),
				"internal/c/c.go": unit("c"),
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

func TestSiblingimportsSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"internal/app/app.go":                  unit("app internal/adapter internal/engine"),
			"internal/engine/e.go":                 unit("engine"),
			"internal/adapter/ad.go":               unit("adapter internal/engine"),
			"internal/app/internal/deep/d.go":      unit("deep internal/adapter"),
			"internal/app/internal/deep/other.go":  unit("deep"),
			"internal/app/internal/deep/other2.go": unit("deep"),
		},
	)
	want := []string{
		"imports sibling adapter, which itself imports sibling engine: siblings form one layer; have the parent's container.go pass adapter in",
		"imports sibling adapter (through app), which itself imports sibling engine: siblings form one layer; have the parent's container.go pass adapter in",
		"imports siblings adapter, engine: wiring siblings together is the parent's job; move it into container.go and inject what this package needs through an interface it declares",
	}
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Snippet,
		)
	}

	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"snippets = %q, want %q",
			got,
			want,
		)
	}
}

func TestSiblingimportsNamesTheOwningModule(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"svc/svc.go":                unit("svc"),
			"svc/internal/app/app.go":   unit("app svc/internal/engine svc/internal/store"),
			"svc/internal/engine/e.go":  unit("engine"),
			"svc/internal/store/s.go":   unit("store"),
			"svc/internal/store/s2.go":  unit("store"),
			"svc/internal/engine/e2.go": unit("engine"),
		},
	)
	isNamed := len(findings) == 1 && strings.Contains(
		findings[0].Snippet,
		"move it into svc/container.go",
	)
	if !isNamed {
		t.Errorf(
			"findings = %v, want one R47 naming svc/container.go",
			findings,
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
	_, err := siblingimport.Siblingimport(
		missing,
		nil,
	)
	if err == nil {
		t.Error("Siblingimports() error = nil, want one for a missing root")
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

	findings, err = siblingimport.Siblingimport(
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

	slices.Sort(got)
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
