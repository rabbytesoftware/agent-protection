package rules

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

//go:embed rules.yaml
var manifest []byte

func Rules() (registry map[string]Rule, err error) {
	err = yaml.Unmarshal(
		manifest,
		&registry,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"parse embedded rules.yaml: %w",
				err,
			)
	}

	return registry,
		nil
}

func Docs() (docs map[string]report.RuleDoc, err error) {
	registry, err := Rules()
	if err != nil {
		return nil,
			err
	}

	docs = map[string]report.RuleDoc{}
	for id, rule := range registry {
		docs[id] = report.RuleDoc{Text: rule.Text, Rationale: rule.Rationale, FixStrategy: rule.FixStrategy, NotAFix: rule.NotAFix}
	}

	return docs,
		nil
}
