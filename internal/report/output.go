package report

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Output is the final report of one run: every finding, plus the
// explanation of each distinct violated rule keyed by rule ID.
type Output struct {
	RuleDocs map[string]RuleDoc `json:"rule_docs"`
	Findings []Finding          `json:"findings"`
}

// JSON renders the report as indented JSON, the machine-readable form
// written to ap-report.json for agents and CI to consume.
func (o Output) JSON() (document string, err error) {
	encoded, err := json.MarshalIndent(
		o,
		"",
		"  ",
	)
	if err != nil {
		return "",
			fmt.Errorf(
				"marshal report: %w",
				err,
			)
	}
	return string(encoded),
		nil
}

// Text renders the report for a terminal: one section per violated rule,
// headed by its explanation, listing every location that violates it,
// followed by a one-line total.
func (o Output) Text() (text string) {
	var builder strings.Builder
	for _, rule := range o.violatedRules() {
		builder.WriteString(o.section(rule))
	}
	fmt.Fprintf(
		&builder,
		"\n%d findings across %d unique rules\n",
		len(o.Findings),
		len(o.RuleDocs),
	)
	return builder.String()
}

func (o Output) violatedRules() (ids []string) {
	for _, finding := range o.Findings {
		ids = append(
			ids,
			finding.Rule,
		)
	}
	return slices.Compact(ids)
}

func (o Output) section(
	rule string,
) (section string) {
	doc := o.RuleDocs[rule]
	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"\n=== %s: %s ===\nWHY: %s\nFIX: %s\n\n",
		rule,
		doc.Text,
		doc.Rationale,
		doc.FixStrategy,
	)
	for _, finding := range o.Findings {
		builder.WriteString(entry(
			finding,
			rule,
		))
	}
	return builder.String()
}

func entry(
	finding Finding,
	rule string,
) (line string) {
	if finding.Rule != rule {
		return ""
	}
	return fmt.Sprintf(
		"  x %s  %s\n",
		finding.Location,
		finding.Snippet,
	)
}
