package typescript_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript"
)

const messages = `[{"filePath":"%s/src/a.ts","messages":[` +
	`{"ruleId":"no-console","line":2,"message":"Unexpected console statement.","severity":1},` +
	`{"ruleId":"custom/rule","line":5,"message":"Custom.","severity":2}]}]`

func TestTypescript(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"eslint.config.js",
		),
		"export default [];\n",
		0o600,
	)
	fakeBun(
		t,
		fmt.Sprintf(
			"printf '%%s' '%s'\nexit 1",
			fmt.Sprintf(
				messages,
				root,
			),
		),
	)
	findings, err := typescript.Typescript(root)
	if err != nil {
		t.Fatalf(
			"Typescript() error = %v",
			err,
		)
	}
	got := summaries(findings)
	want := []string{"TS-ESLINT:no-console src/a.ts:2", "TS-ESLINT:custom/rule src/a.ts:5", "TS-NO-DEBUG-OUTPUT src/a.ts:2"}
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"Typescript() findings = %v, want %v",
			got,
			want,
		)
	}
}

func TestTypescriptEslintFailure(
	t *testing.T,
) {
	fakeBun(
		t,
		"exit 2",
	)
	_, err := typescript.Typescript(t.TempDir())
	if err == nil {
		t.Error("Typescript() error = nil, want an error when eslint prints no report")
	}
}

func fakeBun(
	t *testing.T,
	body string,
) {
	t.Helper()
	directory := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			directory,
			"bun",
		),
		"#!/bin/sh\n"+body+"\n",
		0o755,
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

func writeFile(
	t *testing.T,
	path string,
	content string,
	mode os.FileMode,
) {
	t.Helper()
	err := os.WriteFile(
		path,
		[]byte(content),
		mode,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
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
