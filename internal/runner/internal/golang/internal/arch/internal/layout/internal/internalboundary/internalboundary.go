package internalboundary

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	looseRule     = "R50"
	soleChildRule = "R51"
)

func Internalboundary(
	root string,
) (findings []report.Finding, err error) {
	directories, err := source.Directories(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"internal boundaries: %w",
				err,
			)
	}

	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"internal boundaries: %w",
				err,
			)
	}

	findings, err = layoutFindings(
		root,
		directories,
		paths,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"internal boundaries: %w",
				err,
			)
	}

	return findings,
		nil
}

func layoutFindings(
	root string,
	directories []string,
	paths []string,
) (findings []report.Finding, err error) {
	loose, looseErr := looseFindings(
		root,
		paths,
	)
	boundaries, boundariesErr := boundaryFindings(
		root,
		directories,
		paths,
	)
	return slices.Concat(
			loose,
			boundaries,
		),
		errors.Join(
			looseErr,
			boundariesErr,
		)
}

func looseFindings(
	root string,
	paths []string,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, path := range paths {
		pathFound, pathErr := looseFinding(
			root,
			path,
		)
		findings = append(
			findings,
			pathFound...,
		)
		failures = append(
			failures,
			pathErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func boundaryFindings(
	root string,
	directories []string,
	paths []string,
) (findings []report.Finding, err error) {
	sourceDirectories := packageDirectories(paths)
	failures := []error{}
	for _, boundary := range slices.DeleteFunc(
		directories,
		isNotBoundary,
	) {
		boundaryFound, boundaryErr := soleChildFindings(
			root,
			boundary,
			sourceDirectories,
		)
		findings = append(
			findings,
			boundaryFound...,
		)
		failures = append(
			failures,
			boundaryErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func looseFinding(
	root string,
	path string,
) (findings []report.Finding, err error) {
	loose, err := isLoose(path)
	if err != nil {
		return nil,
			err
	}

	if !loose {
		return nil,
			nil
	}

	location, err := source.RelPath(
		root,
		path,
	)
	return []report.Finding{{
			Rule:     looseRule,
			Location: location,
			Snippet:  "file with behavior directly inside internal/: move it into its own named subdirectory",
		}},
		err
}

func isLoose(
	path string,
) (loose bool, err error) {
	if !source.IsInternalDirectory(filepath.Dir(path)) {
		return false,
			nil
	}

	declarationOnly, err := source.IsDeclarationOnly(path)
	return !declarationOnly,
		err
}

func soleChildFindings(
	root string,
	boundary string,
	sourceDirectories []string,
) (findings []report.Finding, err error) {
	children := childPackages(
		boundary,
		sourceDirectories,
	)
	if len(children) != 1 {
		return nil,
			nil
	}

	location, locationErr := source.RelPath(
		root,
		filepath.Join(
			boundary,
			children[0],
		),
	)
	parent, parentErr := source.RelPath(
		root,
		filepath.Dir(boundary),
	)
	return []report.Finding{{
			Rule:     soleChildRule,
			Location: location,
			Snippet: fmt.Sprintf(
				"only child of this internal/ directory: merge it into %s and drop the internal/ layer",
				parent,
			),
		}},
		errors.Join(
			locationErr,
			parentErr,
		)
}

func packageDirectories(
	paths []string,
) (directories []string) {
	for _, path := range paths {
		directories = append(
			directories,
			filepath.Dir(path),
		)
	}

	slices.Sort(directories)
	return slices.Compact(directories)
}

func childPackages(
	boundary string,
	sourceDirectories []string,
) (children []string) {
	for _, sourceDirectory := range sourceDirectories {
		children = append(
			children,
			childPackage(
				boundary,
				sourceDirectory,
			)...,
		)
	}

	slices.Sort(children)
	return slices.Compact(children)
}

func childPackage(
	boundary string,
	sourceDirectory string,
) (child []string) {
	rel, isBeneath := strings.CutPrefix(
		sourceDirectory,
		boundary+string(filepath.Separator),
	)
	if !isBeneath {
		return nil
	}

	return strings.Split(
		filepath.ToSlash(rel),
		"/",
	)[:1]
}

func isNotBoundary(
	directory string,
) (excluded bool) {
	return !source.IsInternalDirectory(directory)
}
