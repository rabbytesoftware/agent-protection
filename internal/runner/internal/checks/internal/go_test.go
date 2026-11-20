package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckGo_FindsRealViolationsInATempModule(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/m\n\ngo 1.24\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "f.go"), []byte("package m\nfunc F(a int, flag bool) {}\n"), 0o644)

	findings, err := CheckGo(dir)
	if err != nil {
		t.Fatalf("CheckGo: %v", err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "R32" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected CheckGo to surface the real R32 (bool param) violation, got: %v", findings)
	}
}
