package report_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestOutputJSON(
	t *testing.T,
) {
	output := mustReport(
		t,
		[]report.Finding{{Rule: "R32", Location: "a.go:1", Snippet: "x"}},
	)
	document, err := output.JSON()
	if err != nil {
		t.Fatalf(
			"JSON() error = %v",
			err,
		)
	}
	parsed := map[string]json.RawMessage{}
	err = json.Unmarshal(
		[]byte(document),
		&parsed,
	)
	if err != nil {
		t.Fatalf(
			"JSON() is not valid JSON: %v\n%s",
			err,
			document,
		)
	}
	_, hasDocs := parsed["rule_docs"]
	_, hasFindings := parsed["findings"]
	if !hasDocs ||
		!hasFindings {
		t.Errorf(
			"JSON() keys = %v, want rule_docs and findings",
			parsed,
		)
	}
}

func TestOutputText(
	t *testing.T,
) {
	output := mustReport(
		t,
		[]report.Finding{
			{Rule: "R32", Location: "a.go:1", Snippet: "first"},
			{Rule: "R32", Location: "b.go:2", Snippet: "second"},
		},
	)
	text := output.Text()
	headerCount := strings.Count(
		text,
		"=== R32:",
	)
	if headerCount != 1 {
		t.Errorf(
			"rule header appears %d times, want 1:\n%s",
			headerCount,
			text,
		)
	}
	listsBothLocations := strings.Contains(
		text,
		"a.go:1",
	) && strings.Contains(
		text,
		"b.go:2",
	)
	if !listsBothLocations {
		t.Errorf(
			"Text() is missing a finding location:\n%s",
			text,
		)
	}
}

func mustReport(
	t *testing.T,
	findings []report.Finding,
) (output report.Output) {
	t.Helper()
	output, err := report.Report(findings)
	if err != nil {
		t.Fatalf(
			"Report() error = %v",
			err,
		)
	}
	return output
}
