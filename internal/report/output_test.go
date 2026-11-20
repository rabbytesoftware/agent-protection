package report

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuild_DedupesRuleDocsPerUniqueRule(t *testing.T) {
	findings := []Finding{
		{Rule: "R32", Location: "a.go:1", Snippet: "first"},
		{Rule: "R32", Location: "b.go:2", Snippet: "second"},
		{Rule: "R33", Location: "c.go:3", Snippet: "third"},
	}
	out := Build(findings)

	if len(out.Findings) != 3 {
		t.Errorf("expected all 3 findings preserved, got %d", len(out.Findings))
	}
	if len(out.RuleDocs) != 2 {
		t.Errorf("expected exactly 2 unique rule docs (R32, R33), got %d: %v", len(out.RuleDocs), out.RuleDocs)
	}
	if _, ok := out.RuleDocs["R32"]; !ok {
		t.Error("expected R32 rule doc to be present")
	}
}

func TestBuild_UnknownRuleIDIsStillKeptAsAFindingButSkipsDocs(t *testing.T) {
	findings := []Finding{{Rule: "TS-ESLINT:some-plugin-rule", Location: "x.ts:1", Snippet: "msg"}}
	out := Build(findings)

	if len(out.Findings) != 1 {
		t.Fatalf("expected the finding to still be kept, got %d", len(out.Findings))
	}
	if _, ok := out.RuleDocs["TS-ESLINT:some-plugin-rule"]; ok {
		t.Error("expected no rule doc for an unregistered pass-through rule id")
	}
}

func TestBuild_SortsFindingsByRuleThenLocation(t *testing.T) {
	findings := []Finding{
		{Rule: "R33", Location: "b.go:1"},
		{Rule: "R32", Location: "z.go:1"},
		{Rule: "R32", Location: "a.go:1"},
	}
	out := Build(findings)
	want := []string{"a.go:1", "z.go:1", "b.go:1"}
	for i, f := range out.Findings {
		if f.Location != want[i] {
			t.Errorf("Findings[%d].Location = %q, want %q (full order: %v)", i, f.Location, want[i], out.Findings)
		}
	}
}

func TestOutput_JSON_ProducesValidJSON(t *testing.T) {
	out := Build([]Finding{{Rule: "R32", Location: "a.go:1", Snippet: "x"}})
	j, err := out.JSON()
	if err != nil {
		t.Fatalf("JSON(): %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(j), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, j)
	}
	if _, ok := parsed["rule_docs"]; !ok {
		t.Error("expected top-level 'rule_docs' key")
	}
	if _, ok := parsed["findings"]; !ok {
		t.Error("expected top-level 'findings' key")
	}
}

func TestOutput_Text_ShowsFullExplanationOnlyOnceAcrossMultipleOccurrences(t *testing.T) {
	findings := []Finding{
		{Rule: "R32", Location: "a.go:1", Snippet: "first"},
		{Rule: "R32", Location: "b.go:2", Snippet: "second"},
	}
	out := Build(findings)
	text := out.Text()

	occurrences := strings.Count(text, "=== R32:")
	if occurrences != 1 {
		t.Errorf("expected the rule explanation header to appear exactly once, got %d in:\n%s", occurrences, text)
	}
	if !strings.Contains(text, "a.go:1") || !strings.Contains(text, "b.go:2") {
		t.Error("expected both individual finding locations to still be listed")
	}
}
