package checks

import "testing"

func TestEveryChecker_ImplementsTheCheckerInterface(t *testing.T) {
	checkers := map[string]Checker{
		"go":         GoChecker(),
		"rust":       RustChecker(),
		"typescript": TypeScriptChecker(),
	}
	for name, c := range checkers {
		if c == nil {
			t.Errorf("%s checker is nil", name)
		}
	}
}
