package source

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	goExtension             = ".go"
	testSuffix              = "_test.go"
	generatedHeaderMaxBytes = 500
)

// Source parses every non-test, non-generated .go file under root,
// skipping vendor/, .git/ and node_modules/ trees. It fails when a file
// cannot be read or parsed.
func Source(
	root string,
) (files []File, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	paths, err := selected(
		slices.DeleteFunc(
			found.files,
			IsTest,
		),
		handWrittenPath,
	)
	if err != nil {
		return nil,
			err
	}

	return parseAll(
		root,
		paths,
	)
}

// Tests parses every non-generated _test.go file under root that is not
// guarded by an integration build tag. It fails when a file cannot be
// read or parsed.
func Tests(
	root string,
) (files []File, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	paths, err := selected(
		slices.DeleteFunc(
			found.files,
			isNotTest,
		),
		unitTestPath,
	)
	if err != nil {
		return nil,
			err
	}

	return parseAll(
		root,
		paths,
	)
}

// Paths returns every non-test .go file path under root, generated files
// included, for checks that reason about file layout rather than code.
func Paths(
	root string,
) (paths []string, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	return slices.DeleteFunc(
			found.files,
			IsTest,
		),
		nil
}

// Directories returns every directory under root, root included, skipping
// vendor/, .git/ and node_modules/ trees.
func Directories(
	root string,
) (directories []string, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	return found.directories,
		nil
}

// RelPath returns path relative to root. It fails when path cannot be
// made relative to root, such as a relative path under an absolute root.
func RelPath(
	root string,
	path string,
) (rel string, err error) {
	rel, err = filepath.Rel(
		root,
		path,
	)
	if err != nil {
		return "",
			fmt.Errorf(
				"relative path of %s: %w",
				path,
				err,
			)
	}

	return rel,
		nil
}

// Parse parses content as the Go file at path, which findings report
// relative to root.
func Parse(
	root string,
	path string,
	content []byte,
) (file File, err error) {
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		path,
		content,
		parser.ParseComments,
	)
	if err != nil {
		return File{},
			fmt.Errorf(
				"parse %s: %w",
				path,
				err,
			)
	}

	rel, err := RelPath(
		root,
		path,
	)
	if err != nil {
		return File{},
			err
	}

	return File{Path: path, Rel: rel, FileSet: fileSet, Syntax: syntax, Content: content},
		nil
}

// IsTest reports whether path names a _test.go file.
func IsTest(
	path string,
) (isTest bool) {
	return strings.HasSuffix(
		path,
		testSuffix,
	)
}

// IsGenerated reports whether the first bytes of the file at path carry a
// "Code generated" or "DO NOT EDIT" marker. It fails when the file cannot
// be read.
func IsGenerated(
	path string,
) (isGenerated bool, err error) {
	content, err := readFile(path)
	if err != nil {
		return false,
			err
	}

	return isGeneratedContent(content),
		nil
}

func isGeneratedContent(
	content []byte,
) (isGenerated bool) {
	header := string(content[:min(
		len(content),
		generatedHeaderMaxBytes,
	)])
	return strings.Contains(
		header,
		"Code generated",
	) || strings.Contains(
		header,
		"DO NOT EDIT",
	)
}

func walk(
	root string,
) (found *walker, err error) {
	found = &walker{}
	err = filepath.WalkDir(
		root,
		found.visit,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"walk %s: %w",
				root,
				err,
			)
	}

	return found,
		nil
}

func parseAll(
	root string,
	paths []string,
) (files []File, err error) {
	failures := []error{}
	for _, path := range paths {
		file, parseErr := parse(
			root,
			path,
		)
		files = append(
			files,
			file,
		)
		failures = append(
			failures,
			parseErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return files,
		nil
}

func parse(
	root string,
	path string,
) (file File, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return File{},
			fmt.Errorf(
				"read %s: %w",
				path,
				err,
			)
	}

	return Parse(
		root,
		path,
		content,
	)
}

func selected(
	paths []string,
	selector func(string) ([]string, error),
) (kept []string, err error) {
	failures := []error{}
	for _, path := range paths {
		chosen, selectErr := selector(path)
		kept = append(
			kept,
			chosen...,
		)
		failures = append(
			failures,
			selectErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return kept,
		nil
}

func handWrittenPath(
	path string,
) (kept []string, err error) {
	generated, err := IsGenerated(path)
	if err != nil {
		return nil,
			err
	}

	if generated {
		return nil,
			nil
	}

	return []string{path},
		nil
}

func unitTestPath(
	path string,
) (kept []string, err error) {
	content, err := readFile(path)
	if err != nil {
		return nil,
			err
	}

	isExcluded := isGeneratedContent(content) || hasIntegrationTag(content)
	if isExcluded {
		return nil,
			nil
	}

	return []string{path},
		nil
}

func isNotTest(
	path string,
) (excluded bool) {
	return !IsTest(path)
}

func hasIntegrationTag(
	content []byte,
) (tagged bool) {
	return strings.Contains(
		string(content),
		"//go:build",
	) && strings.Contains(
		string(content),
		"integration",
	)
}

func readFile(
	path string,
) (content []byte, err error) {
	content, err = os.ReadFile(path)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				path,
				err,
			)
	}

	return content,
		nil
}
