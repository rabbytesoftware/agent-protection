package eslint_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/eslint"
)

const twoFiles = `[{"filePath":"ROOT/src/a.ts","messages":[{"ruleId":"no-var","line":3,"message":"Unexpected var.","severity":1}]},{"filePath":"ROOT/b.ts","messages":[]}]`

func TestEslint(
	t *testing.T,
) {
	cases := []struct {
		name    string
		output  string
		exit    int
		want    []string
		wantErr bool
	}{
		{
			name:   "a clean report has every file path made relative to root",
			output: twoFiles,
			want:   []string{"src/a.ts", "b.ts"},
		},
		{
			name:   "banner text before the report is skipped",
			output: "Resolving dependencies\n" + twoFiles,
			want:   []string{"src/a.ts", "b.ts"},
		},
		{
			name:   "a non-zero exit still yields the report",
			output: twoFiles,
			exit:   1,
			want:   []string{"src/a.ts", "b.ts"},
		},
		{
			name:    "output without a JSON report is an error",
			output:  "Oops! Something went wrong",
			exit:    2,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		root := t.TempDir()
		fakeBun(
			t,
			script(
				strings.ReplaceAll(
					tc.output,
					"ROOT",
					root,
				),
				tc.exit,
			),
		)
		assertFilePaths(
			t,
			tc.name,
			root,
			tc.want,
			tc.wantErr,
		)
	}
}

func TestEslintPrefersBunInHome(
	t *testing.T,
) {
	root := t.TempDir()
	home := t.TempDir()
	writeExecutable(
		t,
		filepath.Join(
			home,
			".bun",
			"bin",
			"bun",
		),
		script(
			strings.ReplaceAll(
				twoFiles,
				"ROOT",
				root,
			),
			0,
		),
	)
	t.Setenv(
		"HOME",
		home,
	)
	t.Setenv(
		"PATH",
		t.TempDir(),
	)
	assertFilePaths(
		t,
		"bun under ~/.bun/bin is used when bun is not on PATH",
		root,
		[]string{"src/a.ts", "b.ts"},
		false,
	)
}

func TestEslintWithoutBun(
	t *testing.T,
) {
	t.Setenv(
		"HOME",
		t.TempDir(),
	)
	t.Setenv(
		"PATH",
		t.TempDir(),
	)
	_, err := eslint.Eslint(t.TempDir())
	if err == nil {
		t.Error("Eslint() error = nil, want an error when bun is not installed")
	}
}

func script(
	output string,
	exit int,
) (body string) {
	return fmt.Sprintf(
		"printf '%%s' '%s'\nexit %d",
		output,
		exit,
	)
}

func fakeBun(
	t *testing.T,
	body string,
) {
	t.Helper()
	directory := t.TempDir()
	writeExecutable(
		t,
		filepath.Join(
			directory,
			"bun",
		),
		body,
	)
	t.Setenv(
		"HOME",
		t.TempDir(),
	)
	t.Setenv(
		"PATH",
		directory+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func writeExecutable(
	t *testing.T,
	path string,
	body string,
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
		[]byte("#!/bin/sh\n"+body+"\n"),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}

func assertFilePaths(
	t *testing.T,
	name string,
	root string,
	want []string,
	wantErr bool,
) {
	t.Helper()
	results, err := eslint.Eslint(root)
	if (err != nil) != wantErr {
		t.Errorf(
			"%s: error = %v, wantErr %v",
			name,
			err,
			wantErr,
		)
	}

	got := []string{}
	for _, result := range results {
		got = append(
			got,
			result.FilePath,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: file paths = %v, want %v",
			name,
			got,
			want,
		)
	}
}
