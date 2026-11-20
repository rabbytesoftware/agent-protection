package checks

import "testing"

func TestRustChecker_ReturnsNoErrorOnAnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	findings, err := rustChecker{}.Check(dir)
	if err != nil {
		t.Fatalf("rustChecker.Check: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings from an empty dir with no cargo project, got %v", findings)
	}
}
