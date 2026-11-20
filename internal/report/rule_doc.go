package report

// RuleDoc is the explanation attached once per violated rule: what the
// rule says, why it exists, and how to fix a violation of it.
type RuleDoc struct {
	Text        string `json:"text"`
	Rationale   string `json:"rationale"`
	FixStrategy string `json:"fix_strategy"`
}
