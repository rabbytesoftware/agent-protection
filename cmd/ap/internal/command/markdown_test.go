package command

import (
	"path/filepath"
	"testing"
)

func TestMarkdown(
	t *testing.T,
) {
	directory := t.TempDir()
	valid := filepath.Join(
		directory,
		"valid.json",
	)
	invalid := filepath.Join(
		directory,
		"invalid.json",
	)
	writeFile(
		t,
		valid,
		`{"rule_docs":{},"findings":[]}`,
	)
	writeFile(
		t,
		invalid,
		"not json",
	)
	cases := []struct {
		path     string
		wantCode int
	}{
		{path: valid, wantCode: cleanExitCode},
		{path: invalid, wantCode: findingsExitCode},
		{
			path: filepath.Join(
				directory,
				"missing.json",
			),
			wantCode: findingsExitCode,
		},
	}
	for _, tc := range cases {
		assertCode(
			t,
			"Markdown("+tc.path+")",
			realTools().Markdown(tc.path),
			tc.wantCode,
		)
	}
}
