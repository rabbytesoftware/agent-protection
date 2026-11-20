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
				"internal/a/service.go": "package a\nimport \"example.com/m/internal/b\"\nvar _ = b.X\n",
				"internal/b/handler.go": "package b\nfunc X() {}\n",
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

func TestSiblingImports_UnconventionallyNamedWiringFileIsFlaggedAsR47(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// Named "service.go", NOT container.go/wire.go/<pkg>.go -- but it
		// wires 2 siblings together, so it's acting as a composition root.
		"internal/app/service.go": "package app\nimport (\n\t\"example.com/m/internal/engine\"\n\t\"example.com/m/internal/adapter\"\n)\nvar _ = engine.X\nvar _ = adapter.Y\n",
		"internal/engine/e.go":    "package engine\nfunc X() {}\n",
		"internal/adapter/a.go":   "package adapter\nfunc Y() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SiblingImports(dir, files)
	if err != nil {
		t.Fatalf("SiblingImports: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 consolidated R47 finding, got %d: %v", len(got), got)
	}
	if got[0].Rule != "R47" {
		t.Errorf("expected rule R47, got %q", got[0].Rule)
	}
}

func TestSiblingImports_SingleSiblingImportIsStillOrdinaryR16(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// Imports only ONE sibling -- not acting as a composition root,
		// just an ordinary layering violation. Still R16, not R47. Named
		// "service.go"/"handler.go" (not after their own package) so this
		// doesn't accidentally hit the "file named after its own package"
		// composition-root exemption either.
		"internal/a/service.go": "package a\nimport \"example.com/m/internal/b\"\nvar _ = b.X\n",
		"internal/b/handler.go": "package b\nfunc X() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SiblingImports(dir, files)
	if err != nil {
		t.Fatalf("SiblingImports: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R16" {
		t.Errorf("expected exactly 1 R16 finding, got %v", got)
	}
}

func TestSiblingImports_FileNamedAfterOwnPackageIsExempt(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// "config.go" inside package config -- named after its own package,
		// the other half of rule 10's "main free function lives in the
		// file named after the package" convention. Must be exempt, same
		// as container.go/wire.go.
		"internal/config/config.go": "package config\nimport (\n\t\"example.com/m/internal/engine\"\n\t\"example.com/m/internal/adapter\"\n)\nvar _ = engine.X\nvar _ = adapter.Y\n",
		"internal/engine/e.go":      "package engine\nfunc X() {}\n",
		"internal/adapter/a.go":     "package adapter\nfunc Y() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SiblingImports(dir, files)
	if err != nil {
		t.Fatalf("SiblingImports: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings for config.go in package config, got %v", got)
	}
}

func TestIsInsideInternalTree(t *testing.T) {
	testCases := []struct {
		name string
		dir  string
		want bool
	}{
		{name: "absolute path nested under internal", dir: "/a/b/internal/checks/goast", want: true},
		{name: "absolute path is internal itself", dir: "/a/b/internal", want: true},
		{name: "relative path nested under internal -- the ap-check-dot case", dir: "internal/checks", want: true},
		{name: "relative path is internal itself", dir: "internal", want: true},
		{name: "relative path outside internal entirely", dir: "cmd/ap", want: false},
		{name: "absolute path outside internal entirely", dir: "/a/b/cmd/ap", want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isInsideInternalTree(tc.dir)
			if got != tc.want {
				t.Errorf("isInsideInternalTree(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}
