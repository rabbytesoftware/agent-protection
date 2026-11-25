package assertion

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R42"
	discardedFormat = "test %s discards its *testing.T, so it can never fail: name the parameter t and report a wrong result with t.Errorf or t.Fatalf"
	vacuousFormat   = "test %s can never fail: it reaches no t.Error, t.Errorf, t.Fatal, t.Fatalf, t.Fail or t.FailNow, directly or through a helper that receives t; compare the result with the expected value and call t.Errorf when they differ"
)

func Assertion(
	files []source.File,
	tests []source.File,
) (findings []report.Finding) {
	known := catalogOf(slices.Concat(
		files,
		tests,
	))
	proven := provenDeclarations(known)
	for _, file := range tests {
		findings = append(
			findings,
			fileFindings(
				known,
				proven,
				file,
			)...,
		)
	}

	return findings
}

func fileFindings(
	known catalog,
	proven []int,
	file source.File,
) (findings []report.Finding) {
	alias := testingAlias(importedPaths(file.Syntax))
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				known,
				proven,
				file,
				function,
				alias,
			)...,
		)
	}

	return findings
}

func functionFindings(
	known catalog,
	proven []int,
	file source.File,
	function *ast.FuncDecl,
	alias string,
) (findings []report.Finding) {
	isUnproven := isTestFunction(
		function,
		alias,
	) && !slices.Contains(
		proven,
		known.indexOf(function.Body),
	)
	if !isUnproven {
		return nil
	}

	return []report.Finding{finding(
		file,
		function,
	)}
}

func finding(
	file source.File,
	function *ast.FuncDecl,
) (found report.Finding) {
	format := vacuousFormat
	if isDiscarded(function.Type.Params.List[0]) {
		format = discardedFormat
	}

	return report.Finding{
		Rule:     rule,
		Location: file.Location(function),
		Snippet: fmt.Sprintf(
			format,
			function.Name.Name,
		),
	}
}

func isDiscarded(
	parameter *ast.Field,
) (discarded bool) {
	return len(parameter.Names) == 0 || parameter.Names[0].Name == blankImport
}

func isTestFunction(
	function *ast.FuncDecl,
	alias string,
) (test bool) {
	isCandidate := function.Recv == nil && function.Body != nil && function.Type.Params.NumFields() == 1
	if !isCandidate {
		return false
	}

	parameter := function.Type.Params.List[0].Type
	return isNamedTest(
		function.Name.Name,
		testPrefix,
	) && isTestingSelector(
		pointee(parameter),
		alias,
		[]string{testType},
	) || isNamedTest(
		function.Name.Name,
		fuzzPrefix,
	) && isTestingSelector(
		pointee(parameter),
		alias,
		[]string{fuzzType},
	)
}

func pointee(
	expression ast.Expr,
) (inner ast.Expr) {
	star, isStar := expression.(*ast.StarExpr)
	if !isStar {
		return nil
	}

	return star.X
}

func isNamedTest(
	name string,
	prefix string,
) (named bool) {
	rest, isPrefixed := strings.CutPrefix(
		name,
		prefix,
	)
	if !isPrefixed {
		return false
	}

	first, _ := utf8.DecodeRuneInString(rest)
	return rest == "" || !unicode.IsLower(first)
}
