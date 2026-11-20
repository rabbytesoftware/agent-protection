package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/checks"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// Run is AP's top-level dispatcher: detect root's language, hand off to
// the matching checks.Checker, and build the final report.
func Run(root string) (report.Output, error) {
	lang, err := detectLanguage(root)
	if err != nil {
		return report.Output{}, err
	}
	checker, err := checkerFor(lang)
	if err != nil {
		return report.Output{}, err
	}
	findings, err := checker.Check(root)
	if err != nil {
		return report.Output{}, err
	}
	return report.Build(findings), nil
}

// checkerFor returns the checks.Checker for lang.
func checkerFor(lang string) (checks.Checker, error) {
	switch lang {
	case "go":
		return checks.GoChecker(), nil
	case "rust":
		return checks.RustChecker(), nil
	case "typescript":
		return checks.TypeScriptChecker(), nil
	}
	return nil, fmt.Errorf("unsupported language: %s", lang)
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
