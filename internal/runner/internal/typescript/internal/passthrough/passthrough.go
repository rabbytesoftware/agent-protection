package passthrough

import (
	"fmt"
	"os"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/eslint"
)

const (
	rulePrefix = "TS-ESLINT:"
	off        = 0
)

func Passthrough(
	root string,
) (findings []report.Finding, err error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				root,
				err,
			)
	}

	if !slices.ContainsFunc(
		entries,
		isConfig,
	) {
		return nil,
			nil
	}

	results, err := eslint.Eslint(root)
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
	isSilent := message.Severity == off || message.RuleID == ""
	if isSilent {
		return nil
	}

	return []report.Finding{{
		Rule: rulePrefix + message.RuleID,
		Location: fmt.Sprintf(
			"%s:%d",
			path,
			message.Line,
		),
		Snippet: message.Message,
	}}
}

func isConfig(
	entry os.DirEntry,
) (config bool) {
	return slices.Contains(
		configNames(),
		entry.Name(),
	)
}

func configNames() (names []string) {
	return []string{
		"eslint.config.js",
		"eslint.config.mjs",
		"eslint.config.cjs",
		"eslint.config.ts",
		"eslint.config.mts",
		"eslint.config.cts",
		".eslintrc",
		".eslintrc.js",
		".eslintrc.cjs",
		".eslintrc.json",
		".eslintrc.yaml",
		".eslintrc.yml",
	}
}
