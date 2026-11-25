package assertion

import (
	"go/ast"
	"path"
	"slices"
	"strings"
)

const (
	importQuotes  = "\"`"
	blankImport   = "_"
	dotImport     = "."
	pathSeparator = "/"
	versionPrefix = "v"
	digits        = "0123456789"
)

func importedPaths(
	syntax *ast.File,
) (imports map[string]string) {
	imports = map[string]string{}
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(
			spec.Path.Value,
			importQuotes,
		)
		imports[importName(
			spec,
			importPath,
		)] = importPath
	}

	delete(
		imports,
		blankImport,
	)
	delete(
		imports,
		dotImport,
	)
	return imports
}

func importName(
	spec *ast.ImportSpec,
	importPath string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	name = path.Base(importPath)
	if isMajorVersion(name) {
		return path.Base(path.Dir(importPath))
	}

	return name
}

func isMajorVersion(
	element string,
) (version bool) {
	number, isPrefixed := strings.CutPrefix(
		element,
		versionPrefix,
	)
	return isPrefixed && number != "" && strings.Trim(
		number,
		digits,
	) == ""
}

func isLibraryAssertion(
	imports map[string]string,
	call *ast.CallExpr,
	names []string,
) (asserts bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	return isIdent && isAssertionLibrary(imports[qualifier.Name]) && passesTesting(
		call.Args,
		names,
	)
}

func isAssertionLibrary(
	importPath string,
) (library bool) {
	for _, prefix := range assertionLibraries() {
		library = library || importPath == prefix || strings.HasPrefix(
			importPath,
			prefix+pathSeparator,
		)
	}

	return library
}

func passesTesting(
	arguments []ast.Expr,
	names []string,
) (passes bool) {
	for _, argument := range arguments {
		passes = passes || isTestingValue(
			argument,
			names,
		)
	}

	return passes
}

func isTestingValue(
	expression ast.Expr,
	names []string,
) (testing bool) {
	identifier, isIdent := ast.Unparen(expression).(*ast.Ident)
	return isIdent && identifier.Name != blankImport && slices.Contains(
		names,
		identifier.Name,
	)
}

func assertionLibraries() (prefixes []string) {
	return []string{
		"github.com/stretchr/testify",
		"gotest.tools",
		"github.com/matryer/is",
		"github.com/frankban/quicktest",
		"github.com/onsi/gomega",
		"github.com/alecthomas/assert",
		"github.com/maxatome/go-testdeep",
		"github.com/gkampitakis/go-snaps",
		"github.com/sebdah/goldie",
		"github.com/rogpeppe/go-internal/testscript",
		"go.uber.org/goleak",
		"go.uber.org/mock/gomock",
		"github.com/golang/mock/gomock",
		"testing/slogtest",
	}
}
