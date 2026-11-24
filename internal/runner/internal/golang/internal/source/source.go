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
// skipping vendor/, .git/ and node_modules/ trees.
func Source(
	root string,
) (files []File, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	return parseAll(
		root,
		slices.DeleteFunc(
			found.files,
			isNotHandWrittenSource,
		),
	)
}

// Tests parses every non-generated _test.go file under root that is not
// guarded by an integration build tag.
func Tests(
	root string,
) (files []File, err error) {
	found, err := walk(root)
	if err != nil {
		return nil,
			err
	}

	return parseAll(
		root,
		slices.DeleteFunc(
			found.files,
			isNotUnitTest,
		),
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

// RelPath returns path relative to root, or path unchanged when it cannot
// be made relative.
func RelPath(
	root string,
	path string,
) (rel string) {
	rel, err := filepath.Rel(
		root,
		path,
	)
	if err != nil {
		return path
	}

	return rel
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

	return File{Path: path, Rel: RelPath(
			root,
			path,
		), FileSet: fileSet, Syntax: syntax, Content: content},
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
// "Code generated" or "DO NOT EDIT" marker. Unreadable files count as hand
// written so they still get checked and surface their own error.
func IsGenerated(
	path string,
) (isGenerated bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

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

func isNotHandWrittenSource(
	path string,
) (excluded bool) {
	return IsTest(path) || IsGenerated(path)
}

func isNotUnitTest(
	path string,
) (excluded bool) {
	return !IsTest(path) || IsGenerated(path) || hasIntegrationTag(path)
}

func hasIntegrationTag(
	path string,
) (tagged bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return strings.Contains(
		string(content),
		"//go:build",
	) && strings.Contains(
		string(content),
		"integration",
	)
}
