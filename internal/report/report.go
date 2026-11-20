package report

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rabbytesoftware/agent-protection/internal/rules"
)

type Finding struct {
	Rule     string `json:"rule"`
	Location string `json:"location"`
	Snippet  string `json:"snippet"`
}

type RuleDoc struct {
	Text        string `json:"text"`
	Rationale   string `json:"rationale"`
	FixStrategy string `json:"fix_strategy"`
}

type Output struct {
	RuleDocs map[string]RuleDoc `json:"rule_docs"`
	Findings []Finding           `json:"findings"`
}

func Build(findings []Finding) Output {
	docs := map[string]RuleDoc{}
	for _, f := range findings {
		if _, ok := docs[f.Rule]; ok {
			continue
		}
		r, known := rules.Registry[f.Rule]
		if !known {
			continue
		}
		docs[f.Rule] = RuleDoc{Text: r.Text, Rationale: r.Rationale, FixStrategy: r.FixStrategy}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		return findings[i].Location < findings[j].Location
	})
	return Output{RuleDocs: docs, Findings: findings}
}

func (o Output) JSON() (string, error) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}
	return string(b), nil
}

func (o Output) Text() string {
	out := ""
	seen := map[string]bool{}
	for _, f := range o.Findings {
		if !seen[f.Rule] {
			seen[f.Rule] = true
			doc := o.RuleDocs[f.Rule]
			out += fmt.Sprintf("\n=== %s: %s ===\nWHY: %s\nFIX: %s\n\n", f.Rule, doc.Text, doc.Rationale, doc.FixStrategy)
		}
		out += fmt.Sprintf("  x %s  %s\n", f.Location, f.Snippet)
	}
	out += fmt.Sprintf("\n%d findings across %d unique rules\n", len(o.Findings), len(o.RuleDocs))
	return out
}
