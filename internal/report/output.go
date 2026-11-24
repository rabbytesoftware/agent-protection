package report

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type Output struct {
	RuleDocs map[string]RuleDoc `json:"rule_docs"`
	Findings []Finding          `json:"findings"`
}

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
		"\n=== %s: %s ===\nWHY: %s\nFIX: %s\nNOT A FIX: %s\n\n",
		rule,
		doc.Text,
		doc.Rationale,
		doc.FixStrategy,
		doc.NotAFix,
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
