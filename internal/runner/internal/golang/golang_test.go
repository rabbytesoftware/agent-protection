package golang_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang"
)

func TestGolang(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"module example.com/fixture\n\ngo 1.24\n",
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"branch.go",
		),
		"package fixture\n\n// Branch picks a label.\nfunc Branch(x bool) (label string) {\n\tif x {\n\t\treturn \"yes\"\n\t} else {\n\t\treturn \"no\"\n\t}\n}\n",
	)
	findings, err := golang.Golang(root)
	if err != nil {
		t.Fatalf(
			"Golang() error = %v",
			err,
		)
	}
	got := ruleIDs(findings)
	for _, want := range []string{"R01", "R03", "R42"} {
		assertContains(
			t,
			got,
			want,
		)
	}
}

func assertContains(
	t *testing.T,
	got []string,
	want string,
) {
	t.Helper()
	if !slices.Contains(
		got,
		want,
	) {
		t.Errorf(
			"Golang() rules = %v, want %s among them",
			got,
			want,
		)
	}
}

func ruleIDs(
	findings []report.Finding,
) (ids []string) {
	for _, finding := range findings {
		ids = append(
			ids,
			finding.Rule,
		)
	}
	return ids
}

func writeFile(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.WriteFile(
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
