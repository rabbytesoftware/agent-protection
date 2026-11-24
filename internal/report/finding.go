package report

// Finding is one rule violation: the rule ID it breaks, the file and line
// it was found at, and a short snippet describing the offending code.
type Finding struct {
	Rule     string `json:"rule"`
	Location string `json:"location"`
	Snippet  string `json:"snippet"`
}
