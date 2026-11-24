package lint

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	binary          = "golangci-lint"
	buildTagsEnvVar = "AP_GO_BUILD_TAGS"
	configName      = ".ap-lint-config.generated.yml"
	configFileMode  = 0o644
	typecheck       = "typecheck"
	allMatches      = -1
	issuesExitCode  = 1
)

const (
	pathGroup = iota + 1
	lineGroup
	textGroup
	linterGroup
)

// Lint runs golangci-lint over the module at root with Agent Protection's
// generated config, passing the build tags named in AP_GO_BUILD_TAGS, and
// maps every issue from a known linter to its rule ID. It fails when
// golangci-lint cannot start or when type-checking alone failed.
func Lint(
	root string,
) (findings []report.Finding, err error) {
	configPath := filepath.Join(
		root,
		configName,
	)
	err = os.WriteFile(
		configPath,
		[]byte(config),
		configFileMode,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"write lint config: %w",
				err,
			)
	}

	output, runErr := run(
		root,
		configPath,
	)
	err = errors.Join(
		runErr,
		removeConfig(configPath),
	)
	if err != nil {
		return nil,
			err
	}

	return parse(output)
}

func run(
	root string,
	configPath string,
) (output string, err error) {
	command := exec.Command(
		binary,
		arguments(
			configPath,
			os.Getenv(buildTagsEnvVar),
		)...,
	)
	command.Dir = root
	combined, err := command.CombinedOutput()
	var exitError *exec.ExitError
	reportedIssues := errors.As(
		err,
		&exitError,
	) && exitError.ExitCode() == issuesExitCode
	failed := err != nil && !reportedIssues
	if failed {
		return "",
			fmt.Errorf(
				"run %s: %w: %s",
				binary,
				err,
				combined,
			)
	}

	return string(combined),
		nil
}

func arguments(
	configPath string,
	buildTags string,
) (args []string) {
	args = []string{"run", "-c", configPath, "--max-issues-per-linter=0", "--max-same-issues=0", "--tests=false", "./..."}
	if buildTags == "" {
		return args
	}

	return append(
		args,
		"--build-tags="+buildTags,
	)
}

func removeConfig(
	configPath string,
) (err error) {
	err = os.Remove(configPath)
	if err != nil {
		return fmt.Errorf(
			"remove lint config: %w",
			err,
		)
	}

	return nil
}

func parse(
	output string,
) (findings []report.Finding, err error) {
	issues := issuePattern().FindAllStringSubmatch(
		output,
		allMatches,
	)
	onlyTypecheckFailed := strings.Contains(
		output,
		"("+typecheck+")",
	) && !slices.ContainsFunc(
		issues,
		isNotTypecheck,
	)
	if onlyTypecheckFailed {
		return nil,
			fmt.Errorf(
				"golangci-lint could not type-check this module, so no linter ran: if it needs a build tag to compile, set %s and re-run",
				buildTagsEnvVar,
			)
	}

	for _, issue := range issues {
		findings = append(
			findings,
			issueFindings(issue)...,
		)
	}

	return findings,
		nil
}

func issueFindings(
	issue []string,
) (findings []report.Finding) {
	rule, known := linterRules()[issue[linterGroup]]
	if !known {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: issue[pathGroup] + ":" + issue[lineGroup],
		Snippet:  "[" + issue[linterGroup] + "] " + issue[textGroup],
	}}
}

func isNotTypecheck(
	issue []string,
) (notTypecheck bool) {
	return issue[linterGroup] != typecheck
}

func issuePattern() (pattern *regexp.Regexp) {
	return regexp.MustCompile(`(?m)^(.+?\.go):(\d+):\d+: (.+) \(([a-zA-Z0-9_-]+)\)[ \t\r]*$`)
}

func linterRules() (rules map[string]string) {
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
