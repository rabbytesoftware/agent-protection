package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/checks/goast"
	"github.com/rabbytesoftware/agent-protection/internal/checks/rustast"
	"github.com/rabbytesoftware/agent-protection/internal/checks/tsast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const buildTagsEnvVar = "AP_GO_BUILD_TAGS"

// Run is AP's composition root: the one place that imports and wires
// together every per-language checker package, detects root's language,
// and builds the final report. cmd/ap/main.go is a thin CLI shell that
// only parses arguments and calls this.
func Run(root string) (report.Output, error) {
	lang, err := detectLanguage(root)
	if err != nil {
		return report.Output{}, err
	}
	findings, err := runChecks(root, lang)
	if err != nil {
		return report.Output{}, err
	}
	return report.Build(findings), nil
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

// runChecks dispatches to the single checker for lang.
func runChecks(root, lang string) ([]report.Finding, error) {
	switch lang {
	case "go":
		return checkGo(root)
	case "rust":
		return rustast.RustClippy(root), nil
	case "typescript":
		return checkTypeScript(root), nil
	}
	return nil, fmt.Errorf("unsupported language: %s", lang)
}

// checkTypeScript runs both the repo's own ESLint config (pass-through) and
// AP's own universal TS manifest.
func checkTypeScript(root string) []report.Finding {
	all := tsast.TSESLint(root)
	manifestFindings, merr := tsast.TSManifest(root)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "ts manifest: %v\n", merr)
	}
	return append(all, manifestFindings...)
}

// checkGo runs every Go analyzer against root: the pure-AST rule checks,
// the file-level structural checks, and golangci-lint.
func checkGo(root string) ([]report.Finding, error) {
	files, err := goast.WalkGoFiles(root)
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}
	all := checkGoASTRules(root, files)
	all = append(all, checkGoFileLevelRules(root, files)...)
	return all, nil
}

// checkGoASTRules runs every per-file AST-based analyzer.
func checkGoASTRules(root string, files []goast.ParsedFile) []report.Finding {
	var all []report.Finding
	all = append(all, goast.BoolParam(root, files)...)
	all = append(all, goast.CQSMismatch(root, files)...)
	all = append(all, goast.HiddenSideEffect(root, files)...)
	all = append(all, goast.PackageHygiene(root, files)...)
	all = append(all, goast.NoStutter(root, files)...)
	all = append(all, goast.OneTypePerFile(root, files)...)
	all = append(all, goast.NoElse(root, files)...)
	all = append(all, goast.NakedReturn(root, files)...)
	all = append(all, goast.CtxFirstParam(root, files)...)
	all = append(all, goast.FileLineLimit(root, files)...)
	all = append(all, goast.NoInfraInUnitTests(root, files)...)
	return all
}

// checkGoFileLevelRules runs the structural checks that walk the directory
// tree themselves (test pairing, max files per dir, sibling imports) plus
// golangci-lint, logging and continuing past any individual failure.
func checkGoFileLevelRules(root string, files []goast.ParsedFile) []report.Finding {
	var all []report.Finding

	pairingFindings, pairingErr := goast.TestPairing(root)
	if pairingErr != nil {
		fmt.Fprintf(os.Stderr, "test-pairing check: %v\n", pairingErr)
	}
	all = append(all, pairingFindings...)

	dirFindings, dirErr := goast.MaxFilesPerDir(root)
	if dirErr != nil {
		fmt.Fprintf(os.Stderr, "max-files-per-dir check: %v\n", dirErr)
	}
	all = append(all, dirFindings...)

	archFindings, archErr := goast.SiblingImports(root, files)
	if archErr != nil {
		fmt.Fprintf(os.Stderr, "arch check: %v\n", archErr)
	}
	all = append(all, archFindings...)

	lintFindings, lintErr := goast.GolangciLint(root, os.Getenv(buildTagsEnvVar))
	if lintErr != nil {
		fmt.Fprintf(os.Stderr, "golangci-lint: %v\n", lintErr)
	}
	all = append(all, lintFindings...)

	return all
}
