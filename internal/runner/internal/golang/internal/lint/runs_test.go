package lint_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

const taggedScript = `case "$*" in
*--build-tags=*) printf 'tagged.go:1:1: %s (errcheck)\n' "$*" ;;
esac
printf '%s\n' 'shared.go:2:1: Error return value is not checked (errcheck)'
exit 1`

func TestLintRunsEveryBuildTagSet(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		taggedScript,
	)
	root := moduleWith(
		t,
		map[string]string{
			"rate.go":         "//go:build !prod\n\npackage p\n",
			"rate_prod.go":    "//go:build prod\n\npackage p\n",
			"tool.go":         "//go:build ignore\n\npackage main\n",
			"tools.go":        "//go:build tools\n\npackage p\n",
			"windows.go":      "//go:build windows && !darwin && !linux\n\npackage p\n",
			"host.go":         "//go:build darwin || linux || windows\n\npackage p\n",
			"modern.go":       "//go:build go1.21 && !goexperiment.foo\n\npackage p\n",
			"rate_it_test.go": "//go:build integration\n\npackage p\n",
		},
	)
	findings, err := lintSnippets(root)
	isMismatch := err != nil || len(findings) != 2 || !strings.Contains(
		findings[1],
		"--build-tags=prod",
	)
	if isMismatch {
		t.Errorf(
			"Lint() = %v, %v, want the shared finding once and one prod run",
			findings,
			err,
		)
	}
}

func TestLintCombinesEnvironmentTagsWithModuleTags(
	t *testing.T,
) {
	t.Setenv(
		"AP_GO_BUILD_TAGS",
		"integration",
	)
	fakeGolangciLint(
		t,
		taggedScript,
	)
	findings, err := lintSnippets(moduleWith(
		t,
		map[string]string{"rate_prod.go": "//go:build prod && integration\n\npackage p\n"},
	))
	isMismatch := err != nil || len(findings) != 3 || !strings.Contains(
		findings[2],
		"--build-tags=integration,prod",
	)
	if isMismatch {
		t.Errorf(
			"Lint() = %v, %v, want an integration run and an integration,prod run",
			findings,
			err,
		)
	}
}

func TestLintReportsUntypecheckableTagSets(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		`case "$*" in
*--build-tags=prod*) printf '%s\n' 'rate_prod.go:1:1: undefined: x (typecheck)' ;;
esac
exit 1`,
	)
	assertLint(
		t,
		"a tag set that does not compile is an R70 finding",
		moduleWith(
			t,
			map[string]string{"rate_prod.go": "//go:build prod\n\npackage p\n"},
		),
		[]string{"R70 go.mod"},
		false,
	)
}

func TestLintCapsTheTagMatrix(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		`printf '%s\n' "$*" >> "$PWD/runs.log"
exit 0`,
	)
	files := map[string]string{}
	for index := range 9 {
		files[fmt.Sprintf(
			"t%d.go",
			index,
		)] = fmt.Sprintf(
			"//go:build t%d\n\npackage p\n",
			index,
		)
	}

	root := moduleWith(
		t,
		files,
	)
	findings, err := lintSnippets(root)
	runLog := readFixture(
		t,
		filepath.Join(
			root,
			"runs.log",
		),
	)
	isMismatch := err != nil || len(findings) != 1 || !strings.Contains(
		findings[0],
		"needs 10 golangci-lint runs",
	) || strings.Count(
		runLog,
		"\n",
	) != 8
	if isMismatch {
		t.Errorf(
			"Lint() = %v, %v with runs %q, want one R70 finding after 8 runs",
			findings,
			err,
			runLog,
		)
	}
}

func TestLintRejectsMalformedBuildConstraints(
	t *testing.T,
) {
	fakeGolangciLint(
		t,
		"exit 0",
	)
	assertLint(
		t,
		"a malformed build constraint is an error",
		moduleWith(
			t,
			map[string]string{"bad.go": "//go:build prod &&\n\npackage p\n"},
		),
		nil,
		true,
	)
	assertLint(
		t,
		"a missing root is an error",
		filepath.Join(
			t.TempDir(),
			"missing",
		),
		nil,
		true,
	)
}
