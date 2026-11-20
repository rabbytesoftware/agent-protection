package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/checks"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "check" {
		fmt.Fprintln(os.Stderr, "usage: ap check <root-dir>")
		os.Exit(2)
	}
	root := os.Args[2]

	lang, derr := detectLanguage(root)
	if derr != nil {
		fmt.Fprintln(os.Stderr, derr)
		os.Exit(2)
	}

	var all []report.Finding
	var err error
	switch lang {
	case "go":
		all, err = checkGo(root)
	case "rust":
		all = checks.RustClippy(root)
	case "typescript":
		all = checks.TSESLint(root)
		manifestFindings, merr := checks.TSManifest(root)
		if merr != nil {
			fmt.Fprintf(os.Stderr, "ts manifest: %v\n", merr)
		} else {
			all = append(all, manifestFindings...)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
		os.Exit(1)
	}

	out := report.Build(all)

	j, _ := out.JSON()
	if werr := os.WriteFile("ap-report.json", []byte(j), 0o644); werr != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", werr)
	}

	fmt.Print(out.Text())
	if len(out.Findings) > 0 {
		os.Exit(1)
	}
}

// detectLanguage looks only at root itself (one ap process = one component
// folder = one language, matching how this is wired per-repo in CI: a
// separate job runs `ap check .` inside api/, inside desktop/src-tauri/,
// inside web/ -- never one process scanning a whole multi-language monorepo).
func detectLanguage(root string) (string, error) {
	if exists(filepath.Join(root, "go.mod")) {
		return "go", nil
	}
	if exists(filepath.Join(root, "Cargo.toml")) {
		return "rust", nil
	}
	if exists(filepath.Join(root, "package.json")) {
		return "typescript", nil
	}
	return "", fmt.Errorf("no go.mod, Cargo.toml, or package.json found directly in %s -- point ap at the single-language component folder itself", root)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkGo(root string) ([]report.Finding, error) {
	files, err := checks.WalkGoFiles(root)
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}

	var all []report.Finding
	all = append(all, checks.BoolParam(root, files)...)
	all = append(all, checks.CQSMismatch(root, files)...)
	all = append(all, checks.HiddenSideEffect(root, files)...)
	all = append(all, checks.PackageHygiene(root, files)...)
	all = append(all, checks.NoStutter(root, files)...)
	all = append(all, checks.OneTypePerFile(root, files)...)
	all = append(all, checks.NoElse(root, files)...)
	all = append(all, checks.NakedReturn(root, files)...)
	all = append(all, checks.CtxFirstParam(root, files)...)
	all = append(all, checks.FileLineLimit(root, files)...)
	all = append(all, checks.NoInfraInUnitTests(root, files)...)

	if f, ferr := checks.TestPairing(root); ferr == nil {
		all = append(all, f...)
	} else {
		fmt.Fprintf(os.Stderr, "test-pairing check: %v\n", ferr)
	}
	if f, ferr := checks.MaxFilesPerDir(root); ferr == nil {
		all = append(all, f...)
	} else {
		fmt.Fprintf(os.Stderr, "max-files-per-dir check: %v\n", ferr)
	}
	if f, ferr := checks.SiblingImports(root, files); ferr == nil {
		all = append(all, f...)
	} else {
		fmt.Fprintf(os.Stderr, "arch check: %v\n", ferr)
	}
	if f, ferr := checks.GolangciLint(root, os.Getenv("AP_GO_BUILD_TAGS")); ferr == nil {
		all = append(all, f...)
	} else {
		fmt.Fprintf(os.Stderr, "golangci-lint: %v\n", ferr)
	}
	return all, nil
}
