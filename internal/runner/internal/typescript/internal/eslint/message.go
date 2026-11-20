package eslint

// Message is one problem eslint reported in a file: the rule that raised
// it (empty for fatal parse errors), its 1-based line, the message text,
// and its severity (0 off, 1 warning, 2 error).
type Message struct {
	RuleID   string `json:"ruleId"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}
