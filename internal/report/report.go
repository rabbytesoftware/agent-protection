package report

import (
	"cmp"
	"maps"
	"slices"
)

func Report(
	findings []Finding,
	registry map[string]RuleDoc,
) (output Output) {
	sorted := slices.SortedFunc(
		slices.Values(findings),
		compareFindings,
	)
	return Output{RuleDocs: ruleDocs(
		registry,
		sorted,
	), Findings: sorted}
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
	registry map[string]RuleDoc,
	findings []Finding,
) (docs map[string]RuleDoc) {
	docs = map[string]RuleDoc{}
	for _, finding := range findings {
		docs[finding.Rule] = registry[finding.Rule]
	}

	maps.DeleteFunc(
		docs,
		isUnregistered,
	)
	return docs
}

func isUnregistered(
	_ string,
	doc RuleDoc,
) (unregistered bool) {
	return doc.Text == ""
}
