package testpairing

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	missingTestRule = "R42"
	emptySnippet    = "empty source file: delete it; a file exists for the declarations it holds"
	orphanTestRule  = "R43"
	goExtension     = ".go"
	testSuffix      = "_test.go"
)

func Testpairing(
	root string,
) (findings []report.Finding, err error) {
	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"test pairing: %w",
				err,
			)
	}

	tests, err := testPaths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"test pairing: %w",
				err,
			)
	}

	findings, err = pairingFindings(
		root,
		paths,
		tests,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"test pairing: %w",
				err,
			)
	}

	return findings,
		nil
}

func pairingFindings(
	root string,
	paths []string,
	tests []string,
) (findings []report.Finding, err error) {
	untested, untestedErr := untestedSources(
		root,
		paths,
		tests,
	)
	orphans, orphansErr := orphanTests(
		root,
		tests,
		paths,
	)
	return slices.Concat(
			untested,
			orphans,
		),
		errors.Join(
			untestedErr,
			orphansErr,
		)
}

func untestedSources(
	root string,
	paths []string,
	tests []string,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, path := range paths {
		pathFound, pathErr := untestedSource(
			root,
			path,
			tests,
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

func untestedSource(
	root string,
	path string,
	tests []string,
) (findings []report.Finding, err error) {
	syntax, err := handWritten(path)
	if syntax == nil ||
		err != nil {
		return nil,
			err
	}

	if !slices.ContainsFunc(
		syntax.Decls,
		isDeclaration,
	) {
		return emptySourceFindings(
			root,
			path,
		)
	}

	if !slices.ContainsFunc(
		syntax.Decls,
		isFunction,
	) {
		return nil,
			nil
	}

	return missingTestFindings(
		root,
		path,
		tests,
	)
}

func orphanTests(
	root string,
	tests []string,
	paths []string,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, test := range tests {
		testFound, testErr := orphanTestFindings(
			root,
			test,
			paths,
		)
		findings = append(
			findings,
			testFound...,
		)
		failures = append(
			failures,
			testErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func testPaths(
	root string,
) (tests []string, err error) {
	directories, err := source.Directories(root)
	if err != nil {
		return nil,
			err
	}

	failures := []error{}
	for _, directory := range directories {
		found, readErr := directoryTests(directory)
		tests = append(
			tests,
			found...,
		)
		failures = append(
			failures,
			readErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return tests,
		nil
}

func directoryTests(
	directory string,
) (tests []string, err error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				directory,
				err,
			)
	}

	for _, entry := range slices.DeleteFunc(
		entries,
		isNotTestFile,
	) {
		tests = append(
			tests,
			filepath.Join(
				directory,
				entry.Name(),
			),
		)
	}

	return tests,
		nil
}

func missingTestFindings(
	root string,
	path string,
	tests []string,
) (findings []report.Finding, err error) {
	location, err := source.RelPath(
		root,
		path,
	)
	if err != nil {
		return nil,
			err
	}

	test := strings.TrimSuffix(
		path,
		goExtension,
	) + testSuffix
	isPaired := source.IsMock(location) || slices.Contains(
		tests,
		test,
	)
	if isPaired {
		return nil,
			nil
	}

	return []report.Finding{{
			Rule:     missingTestRule,
			Location: location,
			Snippet: fmt.Sprintf(
				"no %s beside it",
				filepath.Base(test),
			),
		}},
		nil
}

func orphanTestFindings(
	root string,
	test string,
	paths []string,
) (findings []report.Finding, err error) {
	candidates := pairedSources(test)
	index := slices.IndexFunc(
		candidates,
		isIn(paths),
	)
	if index >= 0 {
		return testlessFindings(
			root,
			test,
			candidates[index],
		)
	}

	location, err := source.RelPath(
		root,
		test,
	)
	return []report.Finding{{
			Rule:     orphanTestRule,
			Location: location,
			Snippet: fmt.Sprintf(
				"no %s beside it: move these tests into the test file of the source they cover",
				filepath.Base(candidates[0]),
			),
		}},
		err
}

func pairedSources(
	test string,
) (sources []string) {
	stem := strings.TrimSuffix(
		test,
		testSuffix,
	)
	sources = []string{stem + goExtension}
	for _, kind := range testKinds() {
		sources = appendKindSource(
			sources,
			stem,
			kind,
		)
	}

	return sources
}

func appendKindSource(
	sources []string,
	stem string,
	kind string,
) (candidates []string) {
	trimmed, isKind := strings.CutSuffix(
		stem,
		kind,
	)
	if !isKind {
		return sources
	}

	return append(
		sources,
		trimmed+goExtension,
	)
}

func isIn(
	paths []string,
) (contains func(string) bool) {
	return func(
		candidate string,
	) (found bool) {
		return slices.Contains(
			paths,
			candidate,
		)
	}
}

func emptySourceFindings(
	root string,
	path string,
) (findings []report.Finding, err error) {
	location, err := source.RelPath(
		root,
		path,
	)
	return []report.Finding{{
			Rule:     missingTestRule,
			Location: location,
			Snippet:  emptySnippet,
		}},
		err
}

func testlessFindings(
	root string,
	test string,
	paired string,
) (findings []report.Finding, err error) {
	declares, err := declaresTests(test)
	if declares ||
		err != nil {
		return nil,
			err
	}

	location, err := source.RelPath(
		root,
		test,
	)
	return []report.Finding{{
			Rule:     missingTestRule,
			Location: location,
			Snippet: fmt.Sprintf(
				"%s declares no tests for %s",
				filepath.Base(test),
				filepath.Base(paired),
			),
		}},
		err
}

func testKinds() (kinds []string) {
	return []string{"_bench", "_integration"}
}

func isNotTestFile(
	entry os.DirEntry,
) (excluded bool) {
	return entry.IsDir() || !source.IsTest(entry.Name())
}
