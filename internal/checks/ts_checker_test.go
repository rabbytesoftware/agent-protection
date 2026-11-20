package checks

import "testing"

func TestTSChecker_ReturnsNoErrorOnAnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	findings, err := tsChecker{}.Check(dir)
	if err != nil {
		t.Fatalf("tsChecker.Check: %v", err)
	}
	_ = findings
}
