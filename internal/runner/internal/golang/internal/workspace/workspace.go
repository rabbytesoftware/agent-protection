package workspace

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R70"
	goModFile      = "go.mod"
	rootLocation   = "."
	pathSeparator  = "/"
	unknownModule  = "unnamed module"
	nestedTemplate = "nested module %s (%s) is invisible to go build ./... and golangci-lint at the root, and the root imports it as if it were external: keep one module per repository root (merge its packages into the root module under internal/, or move it to its own repository), or list it in a go.work at the root so AP checks it as its own module"
)

func Workspace(
	root string,
	check func(string) ([]report.Finding, error),
) (findings []report.Finding, err error) {
	findings, err = check(root)
	if err != nil {
		return nil,
			err
	}

	nested, err := nestedModules(root)
	if err != nil {
		return nil,
			err
	}

	members, err := workMembers(root)
	if err != nil {
		return nil,
			err
	}

	moduleFindings, err := modulesFindings(
		root,
		nested,
		members,
		check,
	)
	return distinct(slices.Concat(
			findings,
			moduleFindings,
		)),
		err
}

func modulesFindings(
	root string,
	nested []string,
	members []string,
	check func(string) ([]report.Finding, error),
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, module := range nested {
		moduleFindings, moduleErr := nestedFindings(
			root,
			module,
			members,
			check,
		)
		findings = append(
			findings,
			moduleFindings...,
		)
		failures = append(
			failures,
			moduleErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func nestedFindings(
	root string,
	module string,
	members []string,
	check func(string) ([]report.Finding, error),
) (findings []report.Finding, err error) {
	directory := filepath.Join(
		root,
		filepath.FromSlash(module),
	)
	if !slices.Contains(
		members,
		module,
	) {
		return []report.Finding{unlistedFinding(
				module,
				directory,
			)},
			nil
	}

	checked, err := check(directory)
	if err != nil {
		return nil,
			fmt.Errorf(
				"check workspace module %s: %w",
				module,
				err,
			)
	}

	return prefixedFindings(
			module,
			checked,
		),
		nil
}

func prefixedFindings(
	module string,
	checked []report.Finding,
) (findings []report.Finding) {
	for _, finding := range checked {
		finding.Location = prefixed(
			module,
			finding.Location,
		)
		findings = append(
			findings,
			finding,
		)
	}

	return findings
}

func unlistedFinding(
	module string,
	directory string,
) (finding report.Finding) {
	modulePath, err := source.ModulePath(directory)
	if err != nil {
		modulePath = unknownModule
	}

	return report.Finding{
		Rule: rule,
		Location: path.Join(
			module,
			goModFile,
		),
		Snippet: fmt.Sprintf(
			nestedTemplate,
			module,
			modulePath,
		),
	}
}

func prefixed(
	module string,
	location string,
) (joined string) {
	if location == rootLocation {
		return module
	}

	return module + pathSeparator + location
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
