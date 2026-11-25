package commentscan_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan"
)

func TestComments(
	t *testing.T,
) {
	cases := []struct {
		name     string
		excluded []string
		want     []string
	}{
		{name: "every language in the tree is scanned", excluded: nil, want: []string{".env:1", ".gitignore:1", "Dockerfile:1", "Justfile:1", "Makefile:1", "a.go:3", "config.jsonc:2", "config.unknown:1", "go.mod:3", "graph/a.graphqls:1", "infra/a.tf:1", "lib/a.lua:1", "lib/a.rs:1", "migrations/1.sql:1", "proto/a.proto:2", "scripts/a.sh:2", "scripts/deploy:2", "settings.ini:1", "tools/a.py:1", "tools/a.rb:1", "tools/a.toml:1", "web/a.css:1", "web/a.html:1", "web/a.html:2", "web/a.html:3", "web/a.ts:1", "workflow.yml:1"}},
		{name: "an excluded extension is left to its language runner", excluded: []string{".go"}, want: []string{".env:1", ".gitignore:1", "Dockerfile:1", "Justfile:1", "Makefile:1", "config.jsonc:2", "config.unknown:1", "go.mod:3", "graph/a.graphqls:1", "infra/a.tf:1", "lib/a.lua:1", "lib/a.rs:1", "migrations/1.sql:1", "proto/a.proto:2", "scripts/a.sh:2", "scripts/deploy:2", "settings.ini:1", "tools/a.py:1", "tools/a.rb:1", "tools/a.toml:1", "web/a.css:1", "web/a.html:1", "web/a.html:2", "web/a.html:3", "web/a.ts:1", "workflow.yml:1"}},
	}
	root := fixture(t)
	for _, tc := range cases {
		assertComments(
			t,
			tc.name,
			root,
			tc.excluded,
			tc.want,
		)
	}
}

func assertComments(
	t *testing.T,
	name string,
	root string,
	excluded []string,
	want []string,
) {
	t.Helper()
	findings, err := commentscan.Comments(
		root,
		excluded,
	)
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Location,
		)
	}

	slices.Sort(got)
	isWrong := err != nil || !slices.Equal(
		got,
		want,
	)
	if isWrong {
		t.Errorf(
			"%s: Comments() = %v, %v, want %v",
			name,
			got,
			err,
			want,
		)
	}
}

func TestCommentsRejectsMissingRoot(
	t *testing.T,
) {
	_, err := commentscan.Comments(
		filepath.Join(
			t.TempDir(),
			"missing",
		),
		nil,
	)
	if err == nil {
		t.Error("Comments() error = nil, want one for a missing root")
	}
}

func fixture(
	t *testing.T,
) (root string) {
	t.Helper()
	root = t.TempDir()
	files := map[string]string{
		"go.mod":           "module x\n\n// pinned\ngo 1.24\n",
		"a.go":             "package x\n\n// note\n",
		"Dockerfile":       "# base\nFROM alpine\n",
		"Makefile":         "# build\nall:\n",
		"workflow.yml":     "# ci\non: push\n",
		"scripts/a.sh":     "#!/bin/sh\n# hack\n",
		"migrations/1.sql": "-- todo\n",
		"web/a.ts":         "// why\n",
		"lib/a.rs":         "// why\n",
		"tools/a.py":       "# why\n",
		"tools/a.rb":       "# why\n",
		"tools/a.toml":     "# why\n",
		"testdata/a.sh":    "# fixture\n",
		"vendor/x/a.go":    "// vendored\n",
		"notes.txt":        "# not code\n",
		"proto/a.proto":    "syntax = \"proto3\";\n// why\n",
		"graph/a.graphqls": "# why\n",
		"web/a.html":       "<!-- why -->\n<p>{{/* hack */}}</p>\n<script>// x\n</script>\n",
		"web/a.css":        "/* why */\n",
		"infra/a.tf":       "# why\n",
		".env":             "# why\n",
		"settings.ini":     "; why\n",
		"config.jsonc":     "{\n// why\n}\n",
		"lib/a.lua":        "-- why\n",
		"Justfile":         "# why\n",
		".gitignore":       "# why\n",
		"scripts/deploy":   "#!/usr/bin/env node\n// why\n",
		"config.unknown":   "# why\n",
		"LICENSE":          "# not code\n",
	}
	for path, code := range files {
		write(
			t,
			filepath.Join(
				root,
				path,
			),
			code,
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
