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
	orphanTestRule  = "R43"
	goExtension     = ".go"
	testSuffix      = "_test.go"
)

// Testpairing pairs source and test files one to one. Every source file
// with behavior needs an x_test.go beside it, and every test file needs the
// x.go it is named after; x_bench_test.go and x_integration_test.go also
// pair with x.go. Declaration-only, generated and mock sources need no
// test of their own.
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

	return slices.Concat(
			untestedSources(
				root,
				paths,
				tests,
			),
			orphanTests(
				root,
				tests,
				paths,
			),
		),
		nil
}

func untestedSources(
	root string,
	paths []string,
	tests []string,
) (findings []report.Finding) {
	for _, path := range slices.DeleteFunc(
		slices.Clone(paths),
		isExempt,
	) {
		findings = append(
			findings,
			missingTestFindings(
				root,
				path,
				tests,
			)...,
		)
	}

	return findings
}

func orphanTests(
	root string,
	tests []string,
	paths []string,
) (findings []report.Finding) {
	for _, test := range tests {
		findings = append(
			findings,
			orphanTestFindings(
				root,
				test,
				paths,
			)...,
		)
	}

	return findings
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
) (findings []report.Finding) {
	location := source.RelPath(
		root,
		path,
	)
	if source.IsMock(location) {
		return nil
	}

	test := strings.TrimSuffix(
		path,
		goExtension,
	) + testSuffix
	if slices.Contains(
		tests,
		test,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     missingTestRule,
		Location: location,
		Snippet: fmt.Sprintf(
			"no %s beside it",
			filepath.Base(test),
		),
	}}
}

func orphanTestFindings(
	root string,
	test string,
	paths []string,
) (findings []report.Finding) {
	candidates := pairedSources(test)
	if containsAny(
		paths,
		candidates,
	) {
		return nil
	}

	path := candidates[0]

	return []report.Finding{{
		Rule: orphanTestRule,
		Location: source.RelPath(
			root,
			test,
		),
		Snippet: fmt.Sprintf(
			"no %s beside it: move these tests into the test file of the source they cover",
			filepath.Base(path),
		),
	}}
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

func containsAny(
	paths []string,
	candidates []string,
) (found bool) {
	for _, candidate := range candidates {
		found = found || slices.Contains(
			paths,
			candidate,
		)
	}

	return found
}

func testKinds() (kinds []string) {
	return []string{"_bench", "_integration"}
}

func isExempt(
	path string,
) (exempt bool) {
	return source.IsDeclarationOnly(path) || source.IsGenerated(path)
}

func isNotTestFile(
	entry os.DirEntry,
) (excluded bool) {
	return entry.IsDir() || !source.IsTest(entry.Name())
}
