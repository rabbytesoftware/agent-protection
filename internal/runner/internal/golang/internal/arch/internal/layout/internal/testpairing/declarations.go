package testpairing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func declaresTests(
	path string,
) (declares bool, err error) {
	syntax, err := handWritten(path)
	if err != nil {
		return false,
			err
	}

	return syntax == nil || slices.ContainsFunc(
			syntax.Decls,
			isTestFunction,
		),
		nil
}

func handWritten(
	path string,
) (syntax *ast.File, err error) {
	generated, err := source.IsGenerated(path)
	if generated ||
		err != nil {
		return nil,
			err
	}

	syntax, err = parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.SkipObjectResolution,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"parse %s: %w",
				path,
				err,
			)
	}

	return syntax,
		nil
}

func isFunction(
	decl ast.Decl,
) (isFunc bool) {
	_, isFunc = decl.(*ast.FuncDecl)
	return isFunc
}

func isDeclaration(
	decl ast.Decl,
) (declares bool) {
	general, isGeneral := decl.(*ast.GenDecl)
	return !isGeneral || general.Tok != token.IMPORT
}

func isTestFunction(
	decl ast.Decl,
) (isTest bool) {
	function, isFunc := decl.(*ast.FuncDecl)
	isPlainFunction := isFunc && function.Recv == nil
	if !isPlainFunction {
		return false
	}

	for _, prefix := range testPrefixes() {
		isTest = isTest || hasTestPrefix(
			function.Name.Name,
			prefix,
		)
	}

	return isTest
}

func hasTestPrefix(
	name string,
	prefix string,
) (matches bool) {
	rest, hasPrefix := strings.CutPrefix(
		name,
		prefix,
	)
	first, _ := utf8.DecodeRuneInString(rest)
	return hasPrefix && (rest == "" || !unicode.IsLower(first))
}

func testPrefixes() (prefixes []string) {
	return []string{"Test", "Benchmark", "Fuzz", "Example"}
}
