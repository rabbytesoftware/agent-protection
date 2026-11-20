package adapter

import "github.com/rabbytesoftware/agent-protection/internal/report"

// Func adapts a plain function to satisfy an interface shaped like
// checks.Checker -- the same pattern as http.HandlerFunc. Checkers are
// genuinely stateless (rule 11: a stateless type is a package-level
// function, not a struct), so they're functions in internal/pipeline, and
// this is the one small piece of glue that lets a function implement an
// interface without becoming an empty struct with a method.
type Func func(root string) ([]report.Finding, error)

// Check calls the wrapped function.
func (f Func) Check(root string) ([]report.Finding, error) {
	return f(root)
}
