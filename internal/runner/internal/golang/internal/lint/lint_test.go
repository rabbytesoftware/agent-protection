package lint_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lint"
)

func TestLint(
	t *testing.T,
) {
	cases := []struct {
		name    string
		script  string
		want    []string
		wantErr bool
	}{
		{
			name: "issues from known linters map to their rules and unknown linters are dropped",
			script: `printf '%s\n' 'pkg/a.go:3:2: Error return value is not checked (errcheck)' 'pkg/b.go:9:1: cyclomatic complexity 9 of func f is high (gocyclo)' 'pkg/c.go:4:1: f - result 0 is always nil (unparam)' 'pkg/d.go:6:2: ineffectual assignment to err (ineffassign)' 'pkg/e.go:7:2: SA4006: this value of err is never used (staticcheck)'
exit 1`,
			want: []string{"R26 pkg/a.go:3", "PROXY-COMPLEXITY pkg/b.go:9", "R26 pkg/d.go:6", "R26 pkg/e.go:7"},
		},
		{
			name:   "a clean run reports nothing",
			script: "exit 0",
			want:   nil,
		},
		{
			name: "a typecheck error alongside real issues still reports the issues",
			script: `printf '%s\n' 'pkg/a.go:1:1: undefined: x (typecheck)' 'pkg/b.go:2:1: naked return in func f (nakedret)'
exit 1`,
			want: []string{"R04 pkg/b.go:2"},
		},
		{
			name: "a typecheck-only failure is an error, not a clean run",
			script: `printf '%s\n' 'pkg/a.go:1:1: undefined: x (typecheck)'
exit 1`,
			wantErr: true,
		},
		{
			name: "a config or load failure is an error even without issue lines",
			script: `printf '%s\n' 'Error: can not load config'
exit 3`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		fakeGolangciLint(
			t,
			tc.script,
		)
		assertLint(
			t,
			tc.name,
			t.TempDir(),
			tc.want,
			tc.wantErr,
		)
	}
}

func TestLintWithoutGolangciLint(
	t *testing.T,
) {
	t.Setenv(
		"PATH",
		t.TempDir(),
	)
	_, err := lint.Lint(t.TempDir())
	if err == nil {
		t.Error("Lint() error = nil, want an error when golangci-lint is not installed")
	}
}

func TestLintPassesBuildTags(
	t *testing.T,
) {
	t.Setenv(
		"AP_GO_BUILD_TAGS",
		"integration",
	)
	fakeGolangciLint(
		t,
		`printf 'pkg/a.go:1:1: %s (errcheck)\n' "$*"
exit 1`,
	)
	findings, err := lint.Lint(t.TempDir())
	if err != nil {
		t.Fatalf(
			"Lint() error = %v",
			err,
		)
	}

	passesTags := len(findings) == 1 && strings.Contains(
		findings[0].Snippet,
		"--build-tags=integration",
	)
	if !passesTags {
		t.Errorf(
			"Lint() findings = %v, want one carrying --build-tags=integration",
			findings,
		)
	}
}

func TestLintConfigChecksBlankAndTypeAssertionErrors(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		`grep -q 'check-blank: true' "$3" || exit 3
grep -q 'check-type-assertions: true' "$3" || exit 3
printf '%s\n' 'pkg/a.go:5:2: Error return value is not checked (errcheck)'
exit 1`,
	)
	assertLint(
		t,
		"generated config enables blank and type assertion checks",
		t.TempDir(),
		[]string{"R26 pkg/a.go:5"},
		false,
	)
}

func fakeGolangciLint(
	t *testing.T,
	script string,
) {
	t.Helper()
	directory := t.TempDir()
	err := os.WriteFile(
		filepath.Join(
			directory,
			"golangci-lint",
		),
		[]byte("#!/bin/sh\n"+script+"\n"),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"write fake golangci-lint: %v",
			err,
		)
	}

	t.Setenv(
		"PATH",
		directory+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func assertLint(
	t *testing.T,
	name string,
	root string,
	want []string,
	wantErr bool,
) {
	t.Helper()
	findings, err := lint.Lint(root)
	if (err != nil) != wantErr {
		t.Errorf(
			"%s: error = %v, wantErr %v",
			name,
			err,
			wantErr,
		)
	}

	got := summaries(findings)
	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}

	_, err = os.Stat(filepath.Join(
		root,
		".ap-lint-config.generated.yml",
	))
	if !errors.Is(
		err,
		fs.ErrNotExist,
	) {
		t.Errorf(
			"%s: generated config left behind in %s",
			name,
			root,
		)
	}
}

func summaries(
	findings []report.Finding,
) (got []string) {
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	return got
}
