package report

type RuleDoc struct {
	Text        string `json:"text"`
	Rationale   string `json:"rationale"`
	FixStrategy string `json:"fix_strategy"`
	NotAFix     string `json:"not_a_fix"`
}
