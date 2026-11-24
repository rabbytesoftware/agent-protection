package rules

type Rule struct {
	ID          string `yaml:"id"`
	Level       string `yaml:"level"`
	Text        string `yaml:"text"`
	Enforcement string `yaml:"enforcement"`
	Rationale   string `yaml:"rationale"`
	FixStrategy string `yaml:"fix_strategy"`
	NotAFix     string `yaml:"not_a_fix"`
}
