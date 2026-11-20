package rules

// Rule is one entry of the universal manifest: a single mechanically
// checkable rule, the level it belongs to, how it is enforced, and the
// explanation and fix strategy an agent receives when it violates it.
type Rule struct {
	ID          string `yaml:"id"`
	Level       string `yaml:"level"`
	Text        string `yaml:"text"`
	Enforcement string `yaml:"enforcement"`
	Rationale   string `yaml:"rationale"`
	FixStrategy string `yaml:"fix_strategy"`
}
