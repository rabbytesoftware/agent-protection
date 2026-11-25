package entrypoint_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/entrypoint"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

type fixture struct {
	name  string
	files map[string]string
	want  []string
}

const fixtureModule = "example.com/m"

const loopingContainer = "package jobq\n\nfunc New() int {\n\tfor {\n\t}\n}\n"

func TestEntrypointRoles(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name: "container.go of a module is checked",
				files: map[string]string{
					"container.go":         loopingContainer,
					"internal/disk/d.go":   "package disk\n",
					"internal/memory/m.go": "package memory\n",
				},
				want: []string{"R14 container.go:4"},
			},
			{
				name:  "container.go of a public package is checked",
				files: map[string]string{"jobq/container.go": loopingContainer},
				want:  []string{"R14 jobq/container.go:4"},
			},
			{
				name:  "container.go of an internal leaf is not an entry point",
				files: map[string]string{"internal/jobq/container.go": loopingContainer},
				want:  nil,
			},
			{
				name:  "container.go under an example tree passes",
				files: map[string]string{"examples/jobq/container.go": loopingContainer},
				want:  nil,
			},
			{
				name:  "container.go under a mock directory passes",
				files: map[string]string{"jobq/mocks/container.go": loopingContainer},
				want:  nil,
			},
			{
				name:  "unexported and method functions in container.go are left to R14 shape checks",
				files: map[string]string{"container.go": "package jobq\n\nfunc helper() {\n\tfor {\n\t}\n}\n\nfunc (q Q) Run() {\n\tfor {\n\t}\n}\n"},
				want:  nil,
			},
			{
				name:  "other files are not entry points",
				files: map[string]string{"jobq.go": loopingContainer},
				want:  nil,
			},
		},
	)
}

func TestEntrypointRejectsMissingModule(
	t *testing.T,
) {
	_, err := entrypoint.Entrypoint(
		t.TempDir(),
		nil,
	)
	if err == nil {
		t.Error("Entrypoint() error = nil, want one without go.mod")
	}
}

func TestEntrypointRejectsUnreadableRoot(
	t *testing.T,
) {
	root := module(
		t,
		nil,
	)
	locked := filepath.Join(
		root,
		"locked",
	)
	err := os.Mkdir(
		locked,
		0o000,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(unlock(locked))
	_, err = entrypoint.Entrypoint(
		root,
		nil,
	)
	if err == nil {
		t.Error("Entrypoint() error = nil, want one for an unreadable directory")
	}
}

func unlock(
	path string,
) (cleanup func()) {
	return func() {
		_ = os.Chmod(
			path,
			0o755,
		)
	}
}

func assertFixtures(
	t *testing.T,
	fixtures []fixture,
) {
	t.Helper()
	for _, each := range fixtures {
		assertFindings(
			t,
			each.name,
			checked(
				t,
				each.files,
			),
			each.want,
		)
	}
}

func checked(
	t *testing.T,
	files map[string]string,
) (findings []report.Finding) {
	t.Helper()
	return checkedAt(
		t,
		fixtureModule,
		files,
	)
}

func checkedAt(
	t *testing.T,
	modulePath string,
	files map[string]string,
) (findings []report.Finding) {
	t.Helper()
	root := moduleAt(
		t,
		modulePath,
		files,
	)
	parsed, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"parse fixture module: %v",
			err,
		)
	}

	findings, err = entrypoint.Entrypoint(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Entrypoint() error = %v",
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
	return moduleAt(
		t,
		fixtureModule,
		files,
	)
}

func moduleAt(
	t *testing.T,
	modulePath string,
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
		"module "+modulePath+"\n\ngo 1.24\n",
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
	want = append(
		[]string{},
		want...,
	)
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	slices.Sort(got)
	slices.Sort(want)
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

func assertSnippet(
	t *testing.T,
	findings []report.Finding,
	location string,
	snippet string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Location+" "+finding.Snippet,
		)
	}

	if !slices.Contains(
		got,
		location+" "+snippet,
	) {
		t.Errorf(
			"no finding at %s with snippet %q in %v",
			location,
			snippet,
			got,
		)
	}
}
