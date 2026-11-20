package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	testCases := []struct {
		name    string
		marker  string
		want    string
		wantErr bool
	}{
		{name: "go.mod detected", marker: "go.mod", want: "go"},
		{name: "Cargo.toml detected", marker: "Cargo.toml", want: "rust"},
		{name: "package.json detected", marker: "package.json", want: "typescript"},
		{name: "no marker is an error", marker: "", wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.marker != "" {
				if err := os.WriteFile(filepath.Join(dir, tc.marker), nil, 0o644); err != nil {
					t.Fatalf("write marker: %v", err)
				}
			}
			got, err := detectLanguage(dir)
			if tc.wantErr {
				if err == nil {
					t.Error("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("detectLanguage() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "present.txt")
	os.WriteFile(present, nil, 0o644)

	if !exists(present) {
		t.Error("expected exists() to be true for a file that was just written")
	}
	if exists(filepath.Join(dir, "absent.txt")) {
		t.Error("expected exists() to be false for a file that was never written")
	}
}

func TestCheckGo_FindsRealViolationsInATempModule(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/m\n\ngo 1.24\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "f.go"), []byte("package m\nfunc F(a int, flag bool) {}\n"), 0o644)

	findings, err := checkGo(dir)
	if err != nil {
		t.Fatalf("checkGo: %v", err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "R32" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected checkGo to surface the real R32 (bool param) violation, got: %v", findings)
	}
}
