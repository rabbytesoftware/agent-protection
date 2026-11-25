package packagehygiene

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	formatRule        = "R21"
	genericRule       = "R22"
	packageSubject    = "package"
	directorySubject  = "directory"
	mainPackage       = "main"
	testPackageSuffix = "_test"
	rootDirectory     = "."
)

func Packagehygiene(
	files []source.File,
) (findings []report.Finding) {
	packages := map[string]source.File{}
	for _, file := range slices.Backward(slices.DeleteFunc(
		slices.Clone(files),
		isMock,
	)) {
		packages[filepath.Dir(file.Rel)] = file
	}

	for _, directory := range slices.Sorted(maps.Keys(packages)) {
		findings = append(
			findings,
			packageFindings(
				directory,
				packages[directory],
			)...,
		)
	}

	return findings
}

func packageFindings(
	directory string,
	file source.File,
) (findings []report.Finding) {
	name := file.Syntax.Name.Name
	findings = nameFindings(
		packageSubject,
		name,
		file.Location(file.Syntax.Name),
	)
	base := filepath.Base(directory)
	isDirectoryChecked := directory != rootDirectory && name != mainPackage && !strings.HasSuffix(
		name,
		testPackageSuffix,
	) && base != name && !isVersion(base)
	if !isDirectoryChecked {
		return findings
	}

	return slices.Concat(
		findings,
		[]report.Finding{{
			Rule:     formatRule,
			Location: directory,
			Snippet: fmt.Sprintf(
				"directory %q holds package %q: name the directory after the package it holds",
				base,
				name,
			),
		}},
		nameFindings(
			directorySubject,
			base,
			directory,
		),
	)
}

func nameFindings(
	subject string,
	name string,
	location string,
) (findings []report.Finding) {
	for _, check := range checks() {
		findings = append(
			findings,
			checkFindings(
				check,
				subject,
				name,
				location,
			)...,
		)
	}

	return findings
}

func checkFindings(
	check nameCheck,
	subject string,
	name string,
	location string,
) (findings []report.Finding) {
	if !check.violates(name) {
		return nil
	}

	return []report.Finding{{
		Rule:     check.rule,
		Location: location,
		Snippet: fmt.Sprintf(
			check.format,
			subject,
			name,
		),
	}}
}

func isMock(
	file source.File,
) (excluded bool) {
	return source.IsMock(file.Rel)
}
