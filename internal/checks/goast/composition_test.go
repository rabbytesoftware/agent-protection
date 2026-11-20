package goast

import "testing"

func TestCompositionRoot_MissingContainerFileIsFlagged(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/goast/a.go": "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go": "package rustast\nfunc A() {}\n",
		"internal/checks/tsast/a.go": "package tsast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R48" {
		t.Fatalf("expected exactly 1 R48 finding for the missing checks.go, got %v", got)
	}
}

func TestCompositionRoot_ContainerWithInterfaceIsClean(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/checks.go":    "package checks\n\n// Checker runs every registered language check.\ntype Checker interface {\n\tCheck(root string) error\n}\n",
		"internal/checks/goast/a.go":   "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go": "package rustast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings for a container exposing an interface, got %v", got)
	}
}

func TestCompositionRoot_ContainerWithNoInterfaceIsFlagged(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/checks.go":    "package checks\n\nfunc Run() {}\n",
		"internal/checks/goast/a.go":   "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go": "package rustast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R48" {
		t.Fatalf("expected exactly 1 R48 finding for an interface-less container, got %v", got)
	}
}

func TestCompositionRoot_ContainerExportingStructIsFlaggedR13(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/checks.go": "package checks\n\n// Checker runs every registered language check.\ntype Checker interface {\n\tCheck(root string) error\n}\n\n// Runner is a concrete implementation, wrongly exported directly.\ntype Runner struct{}\n",
		"internal/checks/goast/a.go":   "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go": "package rustast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R13" {
		t.Fatalf("expected exactly 1 R13 finding for the exported struct, got %v", got)
	}
}

func TestCompositionRoot_SinglePackageDirectoryNeedsNoContainer(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/goast/a.go": "package goast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings -- only 1 child package, not a composition-root candidate, got %v", got)
	}
}

func TestCompositionRoot_LiteralInternalDirectoryIsExempt(t *testing.T) {
	dir := t.TempDir()
	// internal/ itself has 2+ child packages (runner, report) but "internal"
	// isn't a composable Go package name -- must never be flagged.
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/runner/r.go": "package runner\nfunc A() {}\n",
		"internal/report/r.go": "package report\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings for the literal internal/ directory itself, got %v", got)
	}
}

func TestCompositionRoot_ContainerGoNamingIsAlsoRecognized(t *testing.T) {
	dir := t.TempDir()
	// Named "container.go", not "<dirname>.go" -- the other half of the
	// naming convention R47/isWiringFile already recognizes. Caught as a
	// real bug dogfooding against quiver.core's actual internal/app/
	// composition root, which uses exactly this name.
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/app/container.go": "package app\n\n// AppContainer wires engine and adapter together.\ntype AppContainer interface {\n\tRun() error\n}\n",
		"internal/engine/e.go":      "package engine\nfunc A() {}\n",
		"internal/adapter/a.go":     "package adapter\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected container.go to satisfy the composition-root requirement, got %v", got)
	}
}
