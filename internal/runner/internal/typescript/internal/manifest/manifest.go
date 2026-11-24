package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/eslint"
)

const (
	configName     = ".ap-eslint-config.generated.mjs"
	configFileMode = 0o644
)

// Manifest runs eslint over root with Agent Protection's own TypeScript
// rule set in place of the repo's config, and maps every message from one
// of those rules to its rule ID. The generated config is removed again
// afterwards. It fails when eslint cannot run or report.
func Manifest(
	root string,
) (findings []report.Finding, err error) {
	configPath := filepath.Join(
		root,
		configName,
	)
	err = writeConfig(configPath)
	if err != nil {
		return nil,
			err
	}

	results, runErr := eslint.Eslint(
		root,
		"--config",
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

	for _, result := range results {
		findings = append(
			findings,
			resultFindings(result)...,
		)
	}

	return findings,
		nil
}

func writeConfig(
	configPath string,
) (err error) {
	err = os.WriteFile(
		configPath,
		[]byte(config),
		configFileMode,
	)
	if err != nil {
		return fmt.Errorf(
			"write eslint manifest config: %w",
			err,
		)
	}

	return nil
}

func removeConfig(
	configPath string,
) (err error) {
	err = os.Remove(configPath)
	if err != nil {
		return fmt.Errorf(
			"remove eslint manifest config: %w",
			err,
		)
	}

	return nil
}

func resultFindings(
	result eslint.FileResult,
) (findings []report.Finding) {
	for _, message := range result.Messages {
		findings = append(
			findings,
			messageFindings(
				result.FilePath,
				message,
			)...,
		)
	}

	return findings
}

func messageFindings(
	path string,
	message eslint.Message,
) (findings []report.Finding) {
	rule, known := eslintRules()[message.RuleID]
	if !known {
		return nil
	}

	return []report.Finding{{
		Rule: rule,
		Location: fmt.Sprintf(
			"%s:%d",
			path,
			message.Line,
		),
		Snippet: message.Message,
	}}
}

func eslintRules() (rules map[string]string) {
	return map[string]string{
		"complexity":             "TS-PROXY-COMPLEXITY",
		"max-depth":              "TS-PROXY-COMPLEXITY",
		"max-lines-per-function": "TS-PROXY-COMPLEXITY",
		"max-params":             "TS-PROXY-COMPLEXITY",
		"no-magic-numbers":       "TS-PROXY-MAGIC-NUMBER",
		"no-else-return":         "TS-NO-ELSE",
		"no-lonely-if":           "TS-NO-ELSE",
		"no-console":             "TS-NO-DEBUG-OUTPUT",
		"no-param-reassign":      "TS-NO-PARAM-REASSIGN",
		"no-var":                 "TS-MODERN-SYNTAX",
		"prefer-const":           "TS-MODERN-SYNTAX",
		"eqeqeq":                 "TS-MODERN-SYNTAX",
		"max-classes-per-file":   "TS-ONE-TYPE-PER-FILE",
	}
}
