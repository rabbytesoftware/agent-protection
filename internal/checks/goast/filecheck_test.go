package goast

import "testing"

func TestTestPairing(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "source file with matching test file is clean",
			files: map[string]string{
				"a.go":      "package x\nfunc A() {}\n",
				"a_test.go": "package x\nimport \"testing\"\nfunc TestA(t *testing.T) {}\n",
			},
			want: 0,
		},
		{
			name:  "source file with no test file is flagged",
			files: map[string]string{"a.go": "package x\nfunc A() {}\n"},
			want:  1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tc.files {
				writeFile(t, dir, name, content)
			}
			got, err := TestPairing(dir)
			if err != nil {
				t.Fatalf("TestPairing: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("TestPairing() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestMaxFilesPerDir(t *testing.T) {
	testCases := []struct {
		name      string
		fileCount int
		want      int
	}{
		{name: "8 files is clean", fileCount: 8, want: 0},
		{name: "9 files is flagged", fileCount: 9, want: 1},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			sub := "pkg"
			for i := 0; i < tc.fileCount; i++ {
				name := "f" + string(rune('a'+i)) + ".go"
				writeFileDeep(t, dir, sub, name, "package pkg\n")
			}
			got, err := MaxFilesPerDir(dir)
			if err != nil {
				t.Fatalf("MaxFilesPerDir: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("MaxFilesPerDir() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestTestPairing_DeclarationOnlyFileIsExempt(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "types.go", "package x\ntype Rule struct {\n\tID string\n}\nvar Registry = map[string]Rule{}\n")
	got, err := TestPairing(dir)
	if err != nil {
		t.Fatalf("TestPairing: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("declaration-only file should be exempt, got %d findings: %v", len(got), got)
	}
}
