package source

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const (
	goExtension       = ".go"
	testSuffix        = "_test.go"
	integrationTag    = "integration"
	generatedMarker   = `^// Code generated .* DO NOT EDIT\.$`
	maxEnumeratedTags = 12
	buildDirective    = "//go:build"
)

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
			IsTest,
		),
	)
}

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

func IsTest(
	path string,
) (isTest bool) {
	return strings.HasSuffix(
		path,
		testSuffix,
	)
}

func IsGenerated(
	path string,
) (isGenerated bool, err error) {
	comments, err := headerComments(path)
	return isGeneratedHeader(comments) && hasProvenance(path),
		err
}

func isGeneratedHeader(
	comments []*ast.Comment,
) (isGenerated bool) {
	marker := regexp.MustCompile(generatedMarker)
	for _, comment := range comments {
		isGenerated = isGenerated || marker.MatchString(comment.Text)
	}

	return isGenerated
}

func headerComments(
	path string,
) (comments []*ast.Comment, err error) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.PackageClauseOnly|parser.ParseComments,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read header of %s: %w",
				path,
				err,
			)
	}

	return leadingComments(syntax),
		nil
}

func leadingComments(
	syntax *ast.File,
) (comments []*ast.Comment) {
	for _, group := range syntax.Comments {
		comments = append(
			comments,
			commentsBefore(
				group,
				syntax.Package,
			)...,
		)
	}

	return comments
}

func commentsBefore(
	group *ast.CommentGroup,
	position token.Pos,
) (comments []*ast.Comment) {
	if group.End() >= position {
		return nil
	}

	return group.List
}

func walk(
	root string,
) (found *walker, err error) {
	found = &walker{root: filepath.Clean(root)}
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

	file, err = Parse(
		root,
		path,
		content,
	)
	if err != nil {
		return File{},
			err
	}

	file.IsGenerated = isGeneratedHeader(leadingComments(file.Syntax)) && hasProvenance(path)
	return file,
		nil
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

func unitTestPath(
	path string,
) (kept []string, err error) {
	comments, err := headerComments(path)
	if err != nil {
		return nil,
			err
	}

	tagged, err := hasIntegrationTag(comments)
	if tagged {
		return nil,
			err
	}

	return []string{path},
		err
}

func HandWritten(
	files []File,
) (kept []File) {
	return slices.DeleteFunc(
		slices.Clone(files),
		isGeneratedFile,
	)
}

func isGeneratedFile(
	file File,
) (generated bool) {
	return file.IsGenerated
}

func isNotTest(
	path string,
) (excluded bool) {
	return !IsTest(path)
}

func Load(
	root string,
) (files []File, tests []File, err error) {
	files, err = Source(root)
	if err != nil {
		return nil,
			nil,
			fmt.Errorf(
				"load go sources: %w",
				err,
			)
	}

	tests, err = Tests(root)
	if err != nil {
		return nil,
			nil,
			fmt.Errorf(
				"load go tests: %w",
				err,
			)
	}

	return files,
		tests,
		nil
}
