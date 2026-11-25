package source

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
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
	return isGeneratedHeader(comments),
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

	for _, group := range syntax.Comments {
		comments = append(
			comments,
			commentsBefore(
				group,
				syntax.Package,
			)...,
		)
	}

	return comments,
		nil
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
	comments, err := headerComments(path)
	if err != nil {
		return nil,
			err
	}

	tagged, err := hasIntegrationTag(comments)
	isExcluded := tagged || isGeneratedHeader(comments)
	if isExcluded {
		return nil,
			err
	}

	return []string{path},
		err
}

func isNotTest(
	path string,
) (excluded bool) {
	return !IsTest(path)
}

func hasIntegrationTag(
	comments []*ast.Comment,
) (tagged bool, err error) {
	failures := []error{}
	for _, comment := range slices.DeleteFunc(
		slices.Clone(comments),
		isNotGoBuild,
	) {
		integrationOnly, parseErr := isIntegrationOnly(comment.Text)
		tagged = tagged || integrationOnly
		failures = append(
			failures,
			parseErr,
		)
	}

	return tagged,
		errors.Join(failures...)
}

func isNotGoBuild(
	comment *ast.Comment,
) (excluded bool) {
	return !constraint.IsGoBuild(comment.Text)
}

func isIntegrationOnly(
	line string,
) (integrationOnly bool, err error) {
	expression, err := constraint.Parse(line)
	if err != nil {
		return false,
			fmt.Errorf(
				"parse build constraint %q: %w",
				line,
				err,
			)
	}

	others := slices.DeleteFunc(
		constraintTags(line),
		isIntegrationTag,
	)
	isEnumerable := len(others) <= maxEnumeratedTags
	return isEnumerable && isSatisfiable(
			expression,
			others,
			[]string{integrationTag},
		) && !isSatisfiable(
			expression,
			others,
			nil,
		),
		nil
}

func isSatisfiable(
	expression constraint.Expr,
	undecided []string,
	enabled []string,
) (satisfiable bool) {
	if len(undecided) == 0 {
		return expression.Eval(tagSetter(enabled))
	}

	return isSatisfiable(
		expression,
		undecided[1:],
		slices.Concat(
			enabled,
			undecided[:1],
		),
	) || isSatisfiable(
		expression,
		undecided[1:],
		enabled,
	)
}

func tagSetter(
	enabled []string,
) (isSet func(string) bool) {
	return func(
		tag string,
	) (set bool) {
		return slices.Contains(
			enabled,
			tag,
		)
	}
}

func constraintTags(
	line string,
) (tags []string) {
	tags = strings.FieldsFunc(
		strings.TrimPrefix(
			line,
			buildDirective,
		),
		isNotTagRune,
	)
	slices.Sort(tags)
	return slices.Compact(tags)
}

func isNotTagRune(
	character rune,
) (excluded bool) {
	isTagRune := unicode.IsLetter(character) || unicode.IsDigit(character) || character == '_' || character == '.'
	return !isTagRune
}

func isIntegrationTag(
	tag string,
) (matches bool) {
	return tag == integrationTag
}
