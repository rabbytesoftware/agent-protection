package report

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/rules"
)

// Report sorts findings by rule and then by location, and attaches the
// registered explanation of every distinct violated rule exactly once.
// Findings whose rule is not in the registry, such as linter pass-through
// IDs, are kept but carry no explanation.
func Report(
	findings []Finding,
) (output Output, err error) {
	registry, err := rules.Rules()
	if err != nil {
		return Output{},
			fmt.Errorf(
				"load rule registry: %w",
				err,
			)
	}

	sorted := slices.SortedFunc(
		slices.Values(findings),
		compareFindings,
	)
	return Output{RuleDocs: ruleDocs(
			registry,
			sorted,
		), Findings: sorted},
		nil
}

func compareFindings(
	left Finding,
	right Finding,
) (order int) {
	return cmp.Or(
		cmp.Compare(
			left.Rule,
			right.Rule,
		),
		cmp.Compare(
			left.Location,
			right.Location,
		),
	)
}

func ruleDocs(
	registry map[string]rules.Rule,
	findings []Finding,
) (docs map[string]RuleDoc) {
	violated := map[string]rules.Rule{}
	for _, finding := range findings {
		violated[finding.Rule] = registry[finding.Rule]
	}

	maps.DeleteFunc(
		violated,
		isUnregistered,
	)
	docs = map[string]RuleDoc{}
	for id, rule := range violated {
		docs[id] = RuleDoc{Text: rule.Text, Rationale: rule.Rationale, FixStrategy: rule.FixStrategy}
	}

	return docs
}

func isUnregistered(
	_ string,
	rule rules.Rule,
) (unregistered bool) {
	return rule.ID == ""
}
