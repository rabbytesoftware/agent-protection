package goast

import (
	"os"
	"path/filepath"
	"testing"
)

func writeModule(t *testing.T, dir, modulePath string, files map[string]string) {
	t.Helper()
	writeFile(t, dir, "go.mod", "module "+modulePath+"\n\ngo 1.24\n")
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

func TestSiblingImports(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "sibling business packages importing each other is flagged",
			files: map[string]string{
				"internal/a/a.go": "package a\nimport \"example.com/m/internal/b\"\nvar _ = b.X\n",
				"internal/b/b.go": "package b\nfunc X() {}\n",
			},
			want: 1,
		},
		{
			name: "importing a shared leaf support package is clean",
			files: map[string]string{
				"internal/a/a.go":      "package a\nimport \"example.com/m/internal/errors\"\nvar _ = errors.ErrX\n",
				"internal/errors/e.go": "package errors\nimport \"errors\"\nvar ErrX = errors.New(\"x\")\n",
			},
			want: 0,
		},
		{
			name: "composition root (container.go) importing many siblings is clean",
			files: map[string]string{
				"internal/app/container.go": "package app\nimport (\n\t\"example.com/m/internal/engine\"\n\t\"example.com/m/internal/adapter\"\n)\nvar _ = engine.X\nvar _ = adapter.Y\n",
				"internal/engine/e.go":      "package engine\nfunc X() {}\n",
				"internal/adapter/a.go":     "package adapter\nfunc Y() {}\n",
			},
			want: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeModule(t, dir, "example.com/m", tc.files)
			files, err := WalkGoFiles(dir)
			if err != nil {
				t.Fatalf("WalkGoFiles: %v", err)
			}
			got, err := SiblingImports(dir, files)
			if err != nil {
				t.Fatalf("SiblingImports: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("SiblingImports() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}
