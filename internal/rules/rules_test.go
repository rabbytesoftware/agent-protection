package rules_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/rules"
)

func TestRules(
	t *testing.T,
) {
	registry, err := rules.Rules()
	if err != nil {
		t.Fatalf(
			"Rules() error = %v",
			err,
		)
	}

	for id, rule := range registry {
		assertComplete(
			t,
			id,
			rule,
		)
	}
}

func assertComplete(
	t *testing.T,
	id string,
	rule rules.Rule,
) {
	t.Helper()
	isIncomplete := rule.ID != id || rule.Level == "" || rule.Text == "" || rule.Rationale == "" || rule.FixStrategy == ""
	if isIncomplete {
		t.Errorf(
			"rule %q is incomplete or keyed under the wrong ID: %+v",
			id,
			rule,
		)
	}
}
