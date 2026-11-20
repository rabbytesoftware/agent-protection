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

func TestCompositionRoot_LooseFileBesideContainerIsFlaggedR49(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/checks.go": "package checks\n\n// Checker runs a check.\ntype Checker interface {\n\tCheck(root string) error\n}\n",
		// A loose implementation file sitting right beside the container,
		// instead of nested under internal/checks/internal/ -- exactly the
		// mistake this rule exists to catch.
		"internal/checks/go_checker.go": "package checks\nfunc helper() {}\n",
		"internal/checks/goast/a.go":    "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go":  "package rustast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R49" {
		t.Fatalf("expected exactly 1 R49 finding for the stray file, got %v", got)
	}
}

func TestCompositionRoot_OnlyContainerFilePresentIsClean(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		"internal/checks/checks.go":    "package checks\n\n// Checker runs a check.\ntype Checker interface {\n\tCheck(root string) error\n}\n",
		"internal/checks/checks_test.go": "package checks\nimport \"testing\"\nfunc TestX(t *testing.T) {}\n",
		"internal/checks/goast/a.go":   "package goast\nfunc A() {}\n",
		"internal/checks/rustast/a.go": "package rustast\nfunc A() {}\n",
	})
	got, err := CompositionRoot(dir)
	if err != nil {
		t.Fatalf("CompositionRoot: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no R49 findings when the container's own test file is the only other file present, got %v", got)
	}
}

func TestLooseInternalFile_BehaviorfulFileLooseInInternalIsFlagged(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// golangci.go sitting directly in lint/internal/ -- the exact
		// mistake found this session: a single file, no sibling-package
		// pressure (R48 wouldn't trigger), yet still loose in internal/.
		"internal/lint/internal/golangci.go": "package lint\nfunc Run() {}\n",
	})
	got, err := LooseInternalFile(dir)
	if err != nil {
		t.Fatalf("LooseInternalFile: %v", err)
	}
	if len(got) != 1 || got[0].Rule != "R50" {
		t.Fatalf("expected exactly 1 R50 finding, got %v", got)
	}
}

func TestLooseInternalFile_MultipleBehaviorfulFilesAreEachFlagged(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// func.go, go.go, rust.go, typescript.go all loose in checks/internal/
		// -- the exact mistake found this session.
		"internal/checks/internal/func.go":       "package pipeline\nfunc Adapt() {}\n",
		"internal/checks/internal/go.go":          "package pipeline\nfunc CheckGo() {}\n",
		"internal/checks/internal/rust.go":        "package pipeline\nfunc CheckRust() {}\n",
		"internal/checks/internal/typescript.go":  "package pipeline\nfunc CheckTypeScript() {}\n",
	})
	got, err := LooseInternalFile(dir)
	if err != nil {
		t.Fatalf("LooseInternalFile: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 R50 findings (one per loose file), got %d: %v", len(got), got)
	}
}

func TestLooseInternalFile_DeclarationOnlyFileIsExempt(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// component.go -- a pure shared interface, exactly the case that
		// must stay allowed loose in internal/.
		"internal/golang/internal/component.go": "package component\ntype Analyzer interface {\n\tRun(root string) error\n}\n",
	})
	got, err := LooseInternalFile(dir)
	if err != nil {
		t.Fatalf("LooseInternalFile: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings for a declaration-only file, got %v", got)
	}
}

func TestLooseInternalFile_ProperlyNestedFileIsClean(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "example.com/m", map[string]string{
		// golangci.go now in its own subdirectory under internal/ -- the
		// corrected structure.
		"internal/lint/internal/golangci/golangci.go": "package golangci\nfunc Run() {}\n",
	})
	got, err := LooseInternalFile(dir)
	if err != nil {
		t.Fatalf("LooseInternalFile: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no findings once the file is nested in its own subdirectory, got %v", got)
	}
}
