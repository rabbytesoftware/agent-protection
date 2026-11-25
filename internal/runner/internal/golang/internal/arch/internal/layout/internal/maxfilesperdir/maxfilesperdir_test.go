package maxfilesperdir_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/maxfilesperdir"
)

func TestMaxfilesperdir(
	t *testing.T,
) {
	cases := []struct {
		name      string
		directory string
		count     int
		want      []string
	}{
		{name: "eight files pass", directory: "pkg", count: 8, want: nil},
		{name: "nine files are reported at the directory", directory: "pkg", count: 9, want: []string{"R46 pkg"}},
		{name: "root is capped like any directory", directory: ".", count: 9, want: []string{"R46 ."}},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				tc.directory,
				tc.count,
			),
			tc.want,
		)
	}
}

func checked(
	t *testing.T,
	directory string,
	count int,
) (findings []report.Finding) {
	t.Helper()
	root := t.TempDir()
	for index := range count {
		write(
			t,
			filepath.Join(
				root,
				directory,
				fmt.Sprintf(
					"f%d.go",
					index,
				),
			),
			"package pkg\n",
		)
	}

	write(
		t,
		filepath.Join(
			root,
			directory,
			"f0_test.go",
		),
		"package pkg\n",
	)
	findings, err := maxfilesperdir.Maxfilesperdir(root)
	if err != nil {
		t.Fatalf(
			"Maxfilesperdir() error = %v",
			err,
		)
	}

	return findings
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

func TestMaxfilesperdirRejectsMissingRoot(
	t *testing.T,
) {
	missing := filepath.Join(
		t.TempDir(),
		"missing",
	)
	_, err := maxfilesperdir.Maxfilesperdir(missing)
	if err == nil {
		t.Error("Maxfilesperdir() error = nil, want one for a missing root")
	}
}
