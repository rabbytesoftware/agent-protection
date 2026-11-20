package pipeline

import "testing"

func TestCheckRust_ReturnsNoErrorOnAnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	findings, err := CheckRust(dir)
	if err != nil {
		t.Fatalf("CheckRust: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings from an empty dir with no cargo project, got %v", findings)
	}
}
