package lint

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	maxRuns        = 8
	matrixRule     = "R70"
	moduleLocation = "go.mod"
	setSeparator   = "; "
	untaggedSet    = "(no tags)"
)

func runs(
	root string,
	configPath string,
	sets [][]string,
) (findings []report.Finding, err error) {
	output, err := run(
		root,
		configPath,
		sets[0],
	)
	if err != nil {
		return nil,
			err
	}

	findings, typechecked := parse(output)
	if !typechecked {
		return nil,
			fmt.Errorf(
				"golangci-lint could not type-check this module, so no linter ran: if it needs a build tag to compile, set %s and re-run",
				buildTagsEnvVar,
			)
	}

	tagged, err := taggedRuns(
		root,
		configPath,
		sets[1:min(
			len(sets),
			maxRuns,
		)],
	)
	return append(
			append(
				findings,
				tagged...,
			),
			matrixFindings(sets)...,
		),
		err
}

func taggedRuns(
	root string,
	configPath string,
	sets [][]string,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, tags := range sets {
		tagged, runErr := taggedRun(
			root,
			configPath,
			tags,
		)
		findings = append(
			findings,
			tagged...,
		)
		failures = append(
			failures,
			runErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func taggedRun(
	root string,
	configPath string,
	tags []string,
) (findings []report.Finding, err error) {
	output, err := run(
		root,
		configPath,
		tags,
	)
	if err != nil {
		return nil,
			err
	}

	return taggedFindings(
			output,
			tags,
		),
		nil
}

func taggedFindings(
	output string,
	tags []string,
) (findings []report.Finding) {
	findings, typechecked := parse(output)
	if typechecked {
		return findings
	}

	return []report.Finding{{
		Rule:     matrixRule,
		Location: moduleLocation,
		Snippet: fmt.Sprintf(
			"golangci-lint could not type-check the module with -tags=%s, so the files those build tags select were never linted: every tag combination used in //go:build lines must compile",
			strings.Join(
				tags,
				tagSeparator,
			),
		),
	}}
}

func matrixFindings(
	sets [][]string,
) (findings []report.Finding) {
	if len(sets) <= maxRuns {
		return nil
	}

	return []report.Finding{{
		Rule:     matrixRule,
		Location: moduleLocation,
		Snippet: fmt.Sprintf(
			"type-checking every file needs %d golangci-lint runs with different build tags (%s), more than the %d AP runs: merge or drop build tags so at most %d tag sets select every file",
			len(sets),
			describeSets(sets),
			maxRuns,
			maxRuns,
		),
	}}
}

func describeSets(
	sets [][]string,
) (description string) {
	described := []string{}
	for _, tags := range sets {
		described = append(
			described,
			describeTags(tags),
		)
	}

	return strings.Join(
		described,
		setSeparator,
	)
}

func describeTags(
	tags []string,
) (description string) {
	if len(tags) == 0 {
		return untaggedSet
	}

	return strings.Join(
		tags,
		tagSeparator,
	)
}

func distinct(
	findings []report.Finding,
) (unique []report.Finding) {
	for index, finding := range findings {
		unique = append(
			unique,
			firstOccurrence(
				findings[:index],
				finding,
			)...,
		)
	}

	return unique
}

func firstOccurrence(
	earlier []report.Finding,
	finding report.Finding,
) (kept []report.Finding) {
	if slices.Contains(
		earlier,
		finding,
	) {
		return nil
	}

	return []report.Finding{finding}
}
