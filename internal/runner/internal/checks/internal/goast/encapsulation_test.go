package goast

import "testing"

func TestSingleConsumerExposure_BareFolderWithOneImporterIsFlagged(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// goast sits one internal/ segment deep -- reachable from the
		// WHOLE module -- but only pipeline.go actually imports it. This
		// is exactly the real mistake this rule exists to catch.
		"internal/checks/pipeline.go": "package checks\nimport \"example.com/m/internal/goast\"\nvar _ = goast.X\n",
		"internal/goast/g.go":         "package goast\nfunc X() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SingleConsumerExposure(dir, files)
	if err != nil {
		t.Fatalf("SingleConsumerExposure: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R19" {
		t.Fatalf("expected exactly 1 R19 finding for the exposed single-consumer package, got %v", got)
	}
}

func TestSingleConsumerExposure_ProperlyNestedPackageIsClean(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// goast now sits two internal/ segments deep -- Go's own
		// visibility rule already restricts it to the checks package.
		"internal/checks/pipeline.go": "package checks\nimport \"example.com/m/internal/checks/internal/goast\"\nvar _ = goast.X\n",
		"internal/checks/internal/goast/g.go": "package goast\nfunc X() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SingleConsumerExposure(dir, files)
	if err != nil {
		t.Fatalf("SingleConsumerExposure: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings once the package is nested behind a second internal/ boundary, got %v", got)
	}
}

func TestSingleConsumerExposure_MultipleImportersIsClean(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// shared is used by two different packages -- legitimately shared
		// infrastructure, not a single consumer's private detail.
		"internal/a/a.go":      "package a\nimport \"example.com/m/internal/shared\"\nvar _ = shared.X\n",
		"internal/b/b.go":      "package b\nimport \"example.com/m/internal/shared\"\nvar _ = shared.X\n",
		"internal/shared/s.go": "package shared\nfunc X() {}\n",
	})
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	got, err := SingleConsumerExposure(dir, files)
	if err != nil {
		t.Fatalf("SingleConsumerExposure: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings for a package with 2 distinct importers, got %v", got)
	}
}
