package report

// RuleDoc is the explanation attached once per violated rule: what the
// rule says, why it exists, how to fix a violation of it, and which changes
// silence the finding without fixing it.
type RuleDoc struct {
	Text        string `json:"text"`
	Rationale   string `json:"rationale"`
	FixStrategy string `json:"fix_strategy"`
	NotAFix     string `json:"not_a_fix"`
}
