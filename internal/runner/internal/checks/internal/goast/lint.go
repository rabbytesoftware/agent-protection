package goast

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const generatedConfigFileMode = 0o644


const lintConfig = `version: "2"
linters:
  default: none
  exclusions:
    paths:
      - "docs/"
      - "generated/"
      - "gen/"
  enable:
    - gocyclo
    - gocognit
    - funlen
    - nestif
    - mnd
    - dupl
    - gochecknoglobals
    - gochecknoinits
    - unparam
    - errcheck
    - nakedret
    - interfacebloat
  settings:
    interfacebloat:
      max: 3
    nakedret:
      max-func-lines: 0
    gocyclo:
      min-complexity: 8
    gocognit:
      min-complexity: 12
    funlen:
      lines: 40
      statements: -1
    nestif:
      min-complexity: 2
    dupl:
      threshold: 150
`

func linterToRule() map[string]string {
	return map[string]string{
		"gocyclo":          "PROXY-COMPLEXITY",
		"gocognit":         "PROXY-COMPLEXITY",
		"funlen":           "PROXY-COMPLEXITY",
		"nestif":           "PROXY-COMPLEXITY",
		"mnd":              "PROXY-MAGIC-NUMBER",
		"dupl":             "PROXY-DUPLICATION",
		"gochecknoglobals": "PROXY-GLOBAL-STATE",
		"gochecknoinits":   "R24",
		"errcheck":         "R26",
		"nakedret":         "R04",
		"interfacebloat":   "R28",
	}
}

var lintLineRE = regexp.MustCompile(`^(.+?\.go):(\d+):(\d+): (.+) \(([a-zA-Z0-9_-]+)\)\s*$`)

// GolangciLint shells out to golangci-lint with the Agent Protection config.
// buildTags, if non-empty, is passed through as-is -- AP never guesses or
// hardcodes a project-specific build tag; a repo that needs one (e.g. to
// skip a //go:embed target that only exists after a separate frontend
// build step) supplies it via its own CI wiring (the AP_GO_BUILD_TAGS
// environment variable), not via anything baked into this tool.
func GolangciLint(root, buildTags string) ([]report.Finding, error) {
	cfgPath := filepath.Join(root, ".ap-lint-config.generated.yml")
	if err := os.WriteFile(cfgPath, []byte(lintConfig), generatedConfigFileMode); err != nil {
		return nil, fmt.Errorf("write lint config: %w", err)
	}
	defer func() {
		if rmErr := os.Remove(cfgPath); rmErr != nil {
			fmt.Fprintf(os.Stderr, "cleanup generated config: %v\n", rmErr)
		}
	}()

	runArgs := []string{"run", "-c", cfgPath, "--max-issues-per-linter=0", "--max-same-issues=0", "--tests=false", "./..."}
	if buildTags != "" {
		runArgs = append(runArgs, "--build-tags="+buildTags)
	}
	out := runGolangciLint(root, runArgs)

	// If the ONLY thing golangci-lint reported is a typecheck failure, the
	// whole analysis silently aborted -- reporting zero findings here would
	// be a false "clean," not a real result. AP refuses to guess a fix;
	// it surfaces the failure and tells the caller how to supply the build
	// tag their own repo needs.
	if onlyTypecheckIssues(out) {
		return nil, fmt.Errorf("golangci-lint could not type-check this module (typecheck-only failure) -- if this module needs a build tag to compile, set AP_GO_BUILD_TAGS and re-run; refusing to report zero findings as if clean")
	}

	var findings []report.Finding
	for _, line := range splitLines(string(out)) {
		m := lintLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		linter := m[5]
		rule, known := linterToRule()[linter]
		if !known {
			continue
		}
		findings = append(findings, report.Finding{
			Rule:     rule,
			Location: fmt.Sprintf("%s:%s", m[1], m[2]),
			Snippet:  fmt.Sprintf("[%s] %s", linter, m[4]),
		})
	}
	return findings, nil
}

func runGolangciLint(root string, args []string) string {
	cmd := exec.Command("golangci-lint", args...)
	cmd.Dir = root
	out, _ := cmd.CombinedOutput() // golangci-lint exits non-zero when it finds issues; that's expected
	return string(out)
}

// onlyTypecheckIssues reports whether the ONLY issues reported are
// typecheck failures, meaning no other linter actually got to run.
func onlyTypecheckIssues(out string) bool {
	hasTypecheck := regexp.MustCompile(`\(typecheck\)`).MatchString(out)
	hasOther := false
	for _, line := range splitLines(out) {
		if m := lintLineRE.FindStringSubmatch(line); m != nil && m[5] != "typecheck" {
			hasOther = true
		}
	}
	return hasTypecheck && !hasOther
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
