package passthrough_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/passthrough"
)

const messages = `[{"filePath":"%s/src/a.ts","messages":[` +
	`{"ruleId":"no-console","line":2,"message":"Unexpected console statement.","severity":1},` +
	`{"ruleId":"custom/rule","line":5,"message":"Custom.","severity":2},` +
	`{"ruleId":"no-var","line":7,"message":"Disabled.","severity":0},` +
	`{"ruleId":null,"line":1,"message":"Parsing error.","severity":2}]}]`

func TestPassthrough(
	t *testing.T,
) {
	cases := []struct {
		name string
		file string
		want []string
	}{
		{
			name: "flat config messages from enabled rules are reported under TS-ESLINT",
			file: "eslint.config.js",
			want: []string{"TS-ESLINT:no-console src/a.ts:2", "TS-ESLINT:custom/rule src/a.ts:5"},
		},
		{
			name: "legacy eslintrc config is passed through too",
			file: ".eslintrc.json",
			want: []string{"TS-ESLINT:no-console src/a.ts:2", "TS-ESLINT:custom/rule src/a.ts:5"},
		},
		{
			name: "a repo without its own eslint config is skipped",
			file: "package.json",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertPassthrough(
			t,
			tc.name,
			tc.file,
			tc.want,
		)
	}
}

func TestPassthroughUnreadableRoot(
	t *testing.T,
) {
	_, err := passthrough.Passthrough(filepath.Join(
		t.TempDir(),
		"missing",
	))
	if err == nil {
		t.Error("Passthrough() error = nil, want an error for an unreadable root")
	}
}

func TestPassthroughReportFailure(
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
		"exit 2",
	)
	_, err := passthrough.Passthrough(root)
	if err == nil {
		t.Error("Passthrough() error = nil, want an error when eslint prints no report")
	}
}

func assertPassthrough(
	t *testing.T,
	name string,
	file string,
	want []string,
) {
	t.Helper()
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			file,
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
	findings, err := passthrough.Passthrough(root)
	if err != nil {
		t.Fatalf(
			"%s: Passthrough() error = %v",
			name,
			err,
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
