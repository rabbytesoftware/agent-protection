package tsast

type eslintMessage struct {
	RuleID   string `json:"ruleId"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}
