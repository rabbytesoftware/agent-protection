package manifest_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/manifest"
)

func TestManifest(
	t *testing.T,
) {
	cases := []struct {
		eslintRule string
		want       string
	}{
		{eslintRule: "complexity", want: "TS-PROXY-COMPLEXITY"},
		{eslintRule: "max-depth", want: "TS-PROXY-COMPLEXITY"},
		{eslintRule: "max-lines-per-function", want: "TS-PROXY-COMPLEXITY"},
		{eslintRule: "max-params", want: "TS-PROXY-COMPLEXITY"},
		{eslintRule: "no-magic-numbers", want: "TS-PROXY-MAGIC-NUMBER"},
		{eslintRule: "no-else-return", want: "TS-NO-ELSE"},
		{eslintRule: "no-lonely-if", want: "TS-NO-ELSE"},
		{eslintRule: "no-console", want: "TS-NO-DEBUG-OUTPUT"},
		{eslintRule: "no-param-reassign", want: "TS-NO-PARAM-REASSIGN"},
		{eslintRule: "no-var", want: "TS-MODERN-SYNTAX"},
		{eslintRule: "prefer-const", want: "TS-MODERN-SYNTAX"},
		{eslintRule: "eqeqeq", want: "TS-MODERN-SYNTAX"},
		{eslintRule: "max-classes-per-file", want: "TS-ONE-TYPE-PER-FILE"},
		{eslintRule: "some-plugin/unknown", want: ""},
	}
	for _, tc := range cases {
		assertRules(
			t,
			tc.eslintRule,
			tc.want,
		)
	}
}

func TestManifestReportFailure(
	t *testing.T,
) {
	root := t.TempDir()
	fakeBun(
		t,
		"exit 2",
	)
	_, err := manifest.Manifest(root)
	if err == nil {
		t.Error("Manifest() error = nil, want an error when eslint prints no report")
	}
	assertConfigRemoved(
		t,
		root,
	)
}

func assertRules(
	t *testing.T,
	eslintRule string,
	want string,
) {
	t.Helper()
	root := t.TempDir()
	report := fmt.Sprintf(
		`[{"filePath":"%s/src/a.ts","messages":[{"ruleId":"%s","line":4,"message":"m","severity":1}]}]`,
		root,
		eslintRule,
	)
	fakeBun(
		t,
		fmt.Sprintf(
			"[ -f \"$6\" ] || exit 2\nprintf '%%s' '%s'\nexit 1",
			report,
		),
	)
	findings, err := manifest.Manifest(root)
	if err != nil {
		t.Fatalf(
			"%s: Manifest() error = %v",
			eslintRule,
			err,
		)
	}
	got := strings.Join(
		rules(findings),
		",",
	)
	if got != want {
		t.Errorf(
			"%s: rules = %q, want %q",
			eslintRule,
			got,
			want,
		)
	}
	assertConfigRemoved(
		t,
		root,
	)
}

func assertConfigRemoved(
	t *testing.T,
	root string,
) {
	t.Helper()
	_, err := os.Stat(filepath.Join(
		root,
		".ap-eslint-config.generated.mjs",
	))
	if !errors.Is(
		err,
		fs.ErrNotExist,
	) {
		t.Errorf(
			"generated config left behind in %s",
			root,
		)
	}
}

func fakeBun(
	t *testing.T,
	body string,
) {
	t.Helper()
	directory := t.TempDir()
	err := os.WriteFile(
		filepath.Join(
			directory,
			"bun",
		),
		[]byte("#!/bin/sh\n"+body+"\n"),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"write fake bun: %v",
			err,
		)
	}
	t.Setenv(
		"HOME",
		t.TempDir(),
	)
	t.Setenv(
		"PATH",
		directory+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func rules(
	findings []report.Finding,
) (got []string) {
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule,
		)
	}
	return got
}
