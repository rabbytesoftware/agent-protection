package stateless

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	typeRule          = "R11"
	constructorRule   = "R12"
	constructorPrefix = "New"
)

// Stateless reports, package by package, every struct type with no fields
// that still carries methods, and every New or NewXxx constructor whose
// returns all build such an empty struct. Behavior without state belongs
// in package-level functions, and constructors exist only for real state.
func Stateless(
	files []source.File,
) (findings []report.Finding) {
	packages := map[string][]source.File{}
	for _, file := range files {
		directory := filepath.Dir(file.Path)
		packages[directory] = append(
			packages[directory],
			file,
		)
	}

	for _, directory := range slices.Sorted(maps.Keys(packages)) {
		findings = append(
			findings,
			packageFindings(packages[directory])...,
		)
	}

	return findings
}

func packageFindings(
	files []source.File,
) (findings []report.Finding) {
	empty := emptyStructNames(files)
	receivers := receiverNames(files)
	for _, file := range files {
		findings = append(
			findings,
			typeFindings(
				file,
				receivers,
			)...,
		)
		findings = append(
			findings,
			constructorFindings(
				file,
				empty,
			)...,
		)
	}

	return findings
}

func typeFindings(
	file source.File,
	receivers []string,
) (findings []report.Finding) {
	for _, spec := range emptyStructs(file) {
		findings = append(
			findings,
			specFindings(
				file,
				spec,
				receivers,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	spec *ast.TypeSpec,
	receivers []string,
) (findings []report.Finding) {
	if !slices.Contains(
		receivers,
		spec.Name.Name,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     typeRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"type %s has no fields but carries methods: make them package-level functions named after the package",
			spec.Name.Name,
		),
	}}
}

func constructorFindings(
	file source.File,
	empty []string,
) (findings []report.Finding) {
	constructors := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotConstructor,
	)
	for _, constructor := range constructors {
		findings = append(
			findings,
			statelessConstructorFindings(
				file,
				constructor,
				empty,
			)...,
		)
	}

	return findings
}

func statelessConstructorFindings(
	file source.File,
	constructor *ast.FuncDecl,
	empty []string,
) (findings []report.Finding) {
	returns := source.Nodes[*ast.ReturnStmt](constructor.Body)
	isStateless := len(returns) > 0 && allBuildEmpty(
		returns,
		empty,
	)
	if !isStateless {
		return nil
	}

	return []report.Finding{{
		Rule:     constructorRule,
		Location: file.Location(constructor),
		Snippet: fmt.Sprintf(
			"%s constructs a type with no state: drop the constructor and call package-level functions",
			constructor.Name.Name,
		),
	}}
}

func allBuildEmpty(
	returns []*ast.ReturnStmt,
	empty []string,
) (allEmpty bool) {
	allEmpty = true
	for _, statement := range returns {
		allEmpty = allEmpty && buildsEmpty(
			statement,
			empty,
		)
	}

	return allEmpty
}

func buildsEmpty(
	statement *ast.ReturnStmt,
	empty []string,
) (builds bool) {
	if len(statement.Results) == 0 {
		return false
	}

	literal, isLiteral := withoutAddress(statement.Results[0]).(*ast.CompositeLit)
	return isLiteral && slices.Contains(
		empty,
		typeName(literal.Type),
	)
}

func withoutAddress(
	expression ast.Expr,
) (inner ast.Expr) {
	unary, isUnary := expression.(*ast.UnaryExpr)
	isAddress := isUnary && unary.Op == token.AND
	if !isAddress {
		return expression
	}

	return unary.X
}

func isNotConstructor(
	function *ast.FuncDecl,
) (notConstructor bool) {
	rest, hasPrefix := strings.CutPrefix(
		function.Name.Name,
		constructorPrefix,
	)
	isConstructorName := hasPrefix && (rest == "" || strings.IndexFunc(
		rest,
		unicode.IsUpper,
	) == 0)
	return function.Recv != nil || function.Body == nil || !isConstructorName
}

func emptyStructNames(
	files []source.File,
) (names []string) {
	for _, file := range files {
		names = append(
			names,
			fileEmptyStructNames(file)...,
		)
	}

	return names
}

func fileEmptyStructNames(
	file source.File,
) (names []string) {
	for _, spec := range emptyStructs(file) {
		names = append(
			names,
			spec.Name.Name,
		)
	}

	return names
}

func emptyStructs(
	file source.File,
) (specs []*ast.TypeSpec) {
	return slices.DeleteFunc(
		source.Nodes[*ast.TypeSpec](file.Syntax),
		isNotEmptyStruct,
	)
}

func isNotEmptyStruct(
	spec *ast.TypeSpec,
) (notEmpty bool) {
	structure, isStruct := spec.Type.(*ast.StructType)
	return !isStruct || len(structure.Fields.List) > 0
}

func receiverNames(
	files []source.File,
) (names []string) {
	for _, file := range files {
		names = append(
			names,
			fileReceiverNames(file)...,
		)
	}

	return names
}

func fileReceiverNames(
	file source.File,
) (names []string) {
	methods := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotMethod,
	)
	for _, method := range methods {
		names = append(
			names,
			typeName(method.Recv.List[0].Type),
		)
	}

	return names
}

func isNotMethod(
	function *ast.FuncDecl,
) (notMethod bool) {
	return function.Recv == nil || len(function.Recv.List) == 0
}

func typeName(
	expression ast.Expr,
) (name string) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return typeName(typed.X)
	case *ast.ParenExpr:
		return typeName(typed.X)
	case *ast.IndexExpr:
		return typeName(typed.X)
	case *ast.IndexListExpr:
		return typeName(typed.X)
	}

	return ""
}
