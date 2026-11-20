package pipeline

import "testing"

func TestCheckTypeScript_ReturnsNoErrorOnAnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	findings, err := CheckTypeScript(dir)
	if err != nil {
		t.Fatalf("CheckTypeScript: %v", err)
	}
	_ = findings
}
