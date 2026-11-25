package scan_test

import (
	"path/filepath"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/scan"
)

func TestScanResolvesByShebang(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "a shell shebang picks the shell language", code: "#!/bin/sh\nx\n", want: []string{"deploy:1"}},
		{name: "env with flags and a versioned interpreter", code: "#!/usr/bin/env -S python3.11 -u\nx\n", want: []string{"deploy:1"}},
		{name: "env with an assignment", code: "#!/usr/bin/env LANG=C sh\nx\n", want: []string{"deploy:1"}},
		{name: "an unknown interpreter falls back", code: "#!/usr/bin/perl\nx\n", want: nil},
		{name: "a bare env falls back", code: "#!/usr/bin/env\nx\n", want: nil},
		{name: "an empty shebang falls back", code: "#!\nx\n", want: nil},
		{name: "no shebang falls back", code: "x\n", want: nil},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			scanned(
				t,
				map[string]string{"deploy": tc.code},
				lexeme.Language{Spans: noSpans},
			),
			tc.want,
		)
	}
}

func TestScanFallback(
	t *testing.T,
) {
	cases := []struct {
		name     string
		files    map[string]string
		fallback lexeme.Language
		want     []string
	}{
		{name: "an unrecognized text file is lexed by the fallback", files: map[string]string{"a.weird": "x\n"}, fallback: lexeme.Language{Spans: wholeText}, want: []string{"a.weird:1"}},
		{name: "a binary file is skipped", files: map[string]string{"a.bin": "x\x00y\n"}, fallback: lexeme.Language{Spans: wholeText}, want: nil},
		{name: "without a fallback unrecognized files are skipped", files: map[string]string{"a.weird": "x\n"}, fallback: lexeme.Language{}, want: nil},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			scanned(
				t,
				tc.files,
				tc.fallback,
			),
			tc.want,
		)
	}
}

func scanned(
	t *testing.T,
	files map[string]string,
	fallback lexeme.Language,
) (findings []report.Finding) {
	t.Helper()
	root := t.TempDir()
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

	findings, err := scan.Scan(
		root,
		nil,
		[]lexeme.Language{{Matches: isNote, Spans: wholeText, Interpreters: []string{"sh", "python"}}},
		fallback,
	)
	if err != nil {
		t.Fatalf(
			"Scan() error = %v",
			err,
		)
	}

	return findings
}
