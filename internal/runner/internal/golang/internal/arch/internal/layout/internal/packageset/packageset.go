package packageset

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	siblingRule   = "R58"
	duplicateRule = "R59"
	mainPackage   = "main"
)

func Packageset(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	names, err := packageNames(
		root,
		files,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"package set: %w",
				err,
			)
	}

	directories := slices.Sorted(maps.Keys(names))
	return slices.Concat(
			siblingFindings(directories),
			duplicateFindings(
				directories,
				names,
			),
		),
		nil
}

func packageNames(
	root string,
	files []source.File,
) (names map[string]string, err error) {
	names = map[string]string{}
	failures := []error{}
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isTestFile,
	) {
		directory, relErr := source.RelPath(
			root,
			filepath.Dir(file.Path),
		)
		names[filepath.ToSlash(directory)] = file.Syntax.Name.Name
		failures = append(
			failures,
			relErr,
		)
	}

	return names,
		errors.Join(failures...)
}

func siblingFindings(
	directories []string,
) (findings []report.Finding) {
	for _, directory := range directories {
		findings = append(
			findings,
			prefixFindings(
				directory,
				directories,
			)...,
		)
	}

	return findings
}

func prefixFindings(
	directory string,
	directories []string,
) (findings []report.Finding) {
	prefix := longestPrefixSibling(
		directory,
		directories,
	)
	if prefix == "" {
		return nil
	}

	return []report.Finding{{
		Rule:     siblingRule,
		Location: directory,
		Snippet: fmt.Sprintf(
			"%s and its sibling %s split one concept: merge them or name each for what it provides",
			path.Base(directory),
			prefix,
		),
	}}
}

func longestPrefixSibling(
	directory string,
	directories []string,
) (prefix string) {
	for _, candidate := range directories {
		prefix = closerPrefix(
			prefix,
			directory,
			candidate,
		)
	}

	return prefix
}

func closerPrefix(
	prefix string,
	directory string,
	candidate string,
) (closer string) {
	candidateName := path.Base(candidate)
	isCloser := candidate != directory &&
		path.Dir(candidate) == path.Dir(directory) &&
		len(candidateName) > len(prefix) &&
		strings.HasPrefix(
			path.Base(directory),
			candidateName,
		)
	if !isCloser {
		return prefix
	}

	return candidateName
}

func duplicateFindings(
	directories []string,
	names map[string]string,
) (findings []report.Finding) {
	firstUse := map[string]string{}
	for _, directory := range slices.Backward(directories) {
		firstUse[names[directory]] = directory
	}

	for _, directory := range directories {
		findings = append(
			findings,
			duplicateFinding(
				directory,
				names[directory],
				firstUse[names[directory]],
			)...,
		)
	}

	return findings
}

func duplicateFinding(
	directory string,
	name string,
	first string,
) (findings []report.Finding) {
	isDuplicate := name != mainPackage &&
		first != directory &&
		!source.IsMock(name+"/"+name+".go")
	if !isDuplicate {
		return nil
	}

	return []report.Finding{{
		Rule:     duplicateRule,
		Location: directory,
		Snippet: fmt.Sprintf(
			"package name %s is also used by %s: give each package a name that says what it provides",
			name,
			first,
		),
	}}
}

func isTestFile(
	file source.File,
) (excluded bool) {
	return source.IsTest(file.Path)
}
