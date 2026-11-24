package report_test

import (
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	commentLimit = 65536
	longSnippet  = 100
)

func TestOutputMarkdown(
	t *testing.T,
) {
	cases := []struct {
		findings   []report.Finding
		want       []string
		wantAbsent []string
	}{
		{
			findings:   nil,
			want:       []string{"### Agent Protection: no violations"},
			wantAbsent: []string{"<details>"},
		},
		{
			findings:   []report.Finding{{Rule: "R32", Location: "a.go:1", Snippet: "flag"}},
			want:       []string{"1 violation across 1 rule", "| R32 | 1 |", "#### R32: No boolean flag"},
			wantAbsent: []string{"Showing"},
		},
		{
			findings: []report.Finding{
				{Rule: "R32", Location: "a.go:1", Snippet: "first"},
				{Rule: "R32", Location: "b.go:2", Snippet: "second"},
				{Rule: "golangci:errcheck", Location: "c.go:3", Snippet: "unchecked"},
			},
			want:       []string{"3 violations across 2 rules", "<details>", "</details>", "| R32 | 2 |", "| golangci:errcheck | 1 |", "#### golangci:errcheck\n", "- `b.go:2` `second`"},
			wantAbsent: []string{"Showing"},
		},
		{
			findings:   []report.Finding{{Rule: "R32", Location: "a.go:1", Snippet: "uses `backticks`\nand newlines"}},
			want:       []string{"`uses 'backticks' and newlines`"},
			wantAbsent: []string{"``"},
		},
	}
	for _, tc := range cases {
		assertMarkdown(
			t,
			tc.findings,
			tc.want,
			tc.wantAbsent,
		)
	}
}

func TestOutputMarkdownFitsCommentLimit(
	t *testing.T,
) {
	findings := []report.Finding{}
	for range 2000 {
		findings = append(
			findings,
			report.Finding{
				Rule:     "R32",
				Location: "internal/some/deeply/nested/package/file.go:123",
				Snippet: strings.Repeat(
					"x",
					longSnippet,
				),
			},
		)
	}

	document := mustReport(
		t,
		findings,
	).Markdown()
	isTruncated := len(document) < commentLimit && strings.Contains(
		document,
		"of 2000 findings",
	)
	if !isTruncated {
		t.Errorf(
			"Markdown() is %d bytes, want under %d with a truncation note",
			len(document),
			commentLimit,
		)
	}
}

func assertMarkdown(
	t *testing.T,
	findings []report.Finding,
	want []string,
	wantAbsent []string,
) {
	t.Helper()
	document := mustReport(
		t,
		findings,
	).Markdown()
	for _, fragment := range want {
		assertContains(
			t,
			document,
			fragment,
		)
	}

	for _, fragment := range wantAbsent {
		assertAbsent(
			t,
			document,
			fragment,
		)
	}
}

func assertContains(
	t *testing.T,
	document string,
	fragment string,
) {
	t.Helper()
	if !strings.Contains(
		document,
		fragment,
	) {
		t.Errorf(
			"Markdown() is missing %q:\n%s",
			fragment,
			document,
		)
	}
}

func assertAbsent(
	t *testing.T,
	document string,
	fragment string,
) {
	t.Helper()
	if strings.Contains(
		document,
		fragment,
	) {
		t.Errorf(
			"Markdown() unexpectedly contains %q:\n%s",
			fragment,
			document,
		)
	}
}
