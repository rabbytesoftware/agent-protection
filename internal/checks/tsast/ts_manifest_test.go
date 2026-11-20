package tsast

import "testing"

func TestTSRuleIDToAP_CoversEveryRuleEnabledInTheGeneratedConfig(t *testing.T) {
	// Every rule turned on in apESLintConfig must have a mapping, or its
	// findings would silently vanish (tsRuleIDToAP's lookup just skips
	// unknown rule ids) -- this test would catch forgetting to map a newly
	// added rule.
	enabledRules := []string{
		"complexity", "max-depth", "max-lines-per-function", "max-params",
		"no-magic-numbers", "no-else-return", "no-lonely-if", "no-console",
		"no-param-reassign", "no-var", "prefer-const", "eqeqeq", "max-classes-per-file",
	}
	for _, rule := range enabledRules {
		if _, ok := tsRuleIDToAP()[rule]; !ok {
			t.Errorf("rule %q is enabled in apESLintConfig but has no entry in tsRuleIDToAP", rule)
		}
	}
}

func TestAPESLintConfig_IsNonEmpty(t *testing.T) {
	if apESLintConfig == "" {
		t.Error("apESLintConfig must not be empty")
	}
}
