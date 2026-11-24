package internalboundary

import (
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

// Internalboundary checks every directory literally named internal. It may
// directly hold only declaration-only files, and it must hide at least two
// child packages: a sole child has no sibling to be hidden from, so the
// boundary should be merged into its parent.
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
	return slices.Concat(
			looseFindings(
				root,
				paths,
			),
			boundaryFindings(
				root,
				directories,
				paths,
			),
		),
		nil
}

func looseFindings(
	root string,
	paths []string,
) (findings []report.Finding) {
	for _, path := range slices.DeleteFunc(
		slices.Clone(paths),
		isAllowedInBoundary,
	) {
		findings = append(
			findings,
			looseFinding(
				root,
				path,
			),
		)
	}
	return findings
}

func boundaryFindings(
	root string,
	directories []string,
	paths []string,
) (findings []report.Finding) {
	sourceDirectories := packageDirectories(paths)
	for _, boundary := range slices.DeleteFunc(
		directories,
		isNotBoundary,
	) {
		findings = append(
			findings,
			soleChildFindings(
				root,
				boundary,
				sourceDirectories,
			)...,
		)
	}
	return findings
}

func looseFinding(
	root string,
	path string,
) (finding report.Finding) {
	return report.Finding{
		Rule: looseRule,
		Location: source.RelPath(
			root,
			path,
		),
		Snippet: "file with behavior directly inside internal/: move it into its own named subdirectory",
	}
}

func soleChildFindings(
	root string,
	boundary string,
	sourceDirectories []string,
) (findings []report.Finding) {
	children := childPackages(
		boundary,
		sourceDirectories,
	)
	if len(children) != 1 {
		return nil
	}
	return []report.Finding{{
		Rule: soleChildRule,
		Location: source.RelPath(
			root,
			filepath.Join(
				boundary,
				children[0],
			),
		),
		Snippet: fmt.Sprintf(
			"only child of this internal/ directory: merge it into %s and drop the internal/ layer",
			source.RelPath(
				root,
				filepath.Dir(boundary),
			),
		),
	}}
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

func isAllowedInBoundary(
	path string,
) (allowed bool) {
	return !source.IsInternalDirectory(filepath.Dir(path)) || source.IsDeclarationOnly(path)
}

func isNotBoundary(
	directory string,
) (excluded bool) {
	return !source.IsInternalDirectory(directory)
}
