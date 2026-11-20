package checks

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/checks/goast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const buildTagsEnvVar = "AP_GO_BUILD_TAGS"

type goChecker struct{}

func (goChecker) Check(root string) ([]report.Finding, error) {
	files, err := goast.WalkGoFiles(root)
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}
	all := goASTFindings(root, files)
	all = append(all, goFileLevelFindings(root, files)...)
	return all, nil
}

// goASTFindings runs every per-file AST-based analyzer.
func goASTFindings(root string, files []goast.ParsedFile) []report.Finding {
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

// goFileLevelFindings runs the structural checks that walk the directory
// tree themselves (test pairing, max files per dir, sibling imports,
// composition roots) plus golangci-lint, logging and continuing past any
// individual failure.
func goFileLevelFindings(root string, files []goast.ParsedFile) []report.Finding {
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

	containerFindings, containerErr := goast.CompositionRoot(root)
	if containerErr != nil {
		fmt.Fprintf(os.Stderr, "composition-root check: %v\n", containerErr)
	}
	all = append(all, containerFindings...)

	lintFindings, lintErr := goast.GolangciLint(root, os.Getenv(buildTagsEnvVar))
	if lintErr != nil {
		fmt.Fprintf(os.Stderr, "golangci-lint: %v\n", lintErr)
	}
	all = append(all, lintFindings...)

	return all
}
