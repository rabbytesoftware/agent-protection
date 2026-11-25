package droppederror

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	errorName     = "err"
	errorSuffix   = "Err"
	nilName       = "nil"
	recoverName   = "recover"
	literalName   = "literal"
	ignoredFormat = "func %s takes an error parameter it never reads: a reporter that accepts an error must record, wrap or return it; drop the parameter or handle the error"
)

func isFailure(
	name *ast.Ident,
	callees map[string]int,
) (failure bool) {
	return name.Name == errorName || strings.HasSuffix(
		name.Name,
		errorSuffix,
	) || isDeclaredFailure(
		name,
		callees,
	)
}

func isDeclaredFailure(
	name *ast.Ident,
	callees map[string]int,
) (failure bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.Field:
		return isErrorType(declaration.Type)
	case *ast.ValueSpec:
		return isErrorType(declaration.Type) || isBoundToFailure(
			name.Name,
			identifiers(declaration.Names),
			declaration.Values,
			callees,
		)
	case *ast.AssignStmt:
		return isBoundToFailure(
			name.Name,
			declaration.Lhs,
			declaration.Rhs,
			callees,
		)
	}

	return false
}

func isBoundToFailure(
	name string,
	targets []ast.Expr,
	values []ast.Expr,
	callees map[string]int,
) (failure bool) {
	if isRecoveredBinding(
		name,
		targets,
		values,
	) {
		return true
	}

	index := slices.Index(
		identifierNames(targets),
		name,
	)
	isLastResult := index >= 0 && index == len(targets)-1 && len(values) == 1
	return isLastResult && resultsOf(
		values[0],
		callees,
	) == len(targets)
}

func isRecoveredBinding(
	name string,
	targets []ast.Expr,
	values []ast.Expr,
) (recovered bool) {
	index := slices.Index(
		identifierNames(targets),
		name,
	)
	return index >= 0 && len(values) == len(targets) && isRecoverCall(values[index])
}

func isCaughtError(
	name *ast.Ident,
	callees map[string]int,
) (caught bool) {
	return isFailure(
		name,
		callees,
	) && !isRecovered(name)
}

func isRecovered(
	name *ast.Ident,
) (recovered bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.AssignStmt:
		return isRecoveredBinding(
			name.Name,
			declaration.Lhs,
			declaration.Rhs,
		)
	case *ast.ValueSpec:
		return isRecoveredBinding(
			name.Name,
			identifiers(declaration.Names),
			declaration.Values,
		)
	}

	return false
}

func identifiers(
	names []*ast.Ident,
) (expressions []ast.Expr) {
	for _, name := range names {
		expressions = append(
			expressions,
			name,
		)
	}

	return expressions
}

func identifierNames(
	expressions []ast.Expr,
) (names []string) {
	for _, expression := range expressions {
		names = append(
			names,
			identifierName(expression),
		)
	}

	return names
}

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return identifier.Name
}

func calleeName(
	expression ast.Expr,
) (name string) {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return ""
	}

	switch callee := call.Fun.(type) {
	case *ast.Ident:
		return callee.Name
	case *ast.SelectorExpr:
		return callee.Sel.Name
	}

	return ""
}

func resultsOf(
	expression ast.Expr,
	callees map[string]int,
) (count int) {
	count, known := callees[qualifiedName(expression)]
	if known {
		return count
	}

	return callees[calleeName(expression)]
}

func qualifiedName(
	expression ast.Expr,
) (name string) {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return ""
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return qualifier.Name + qualifierSeparator + selector.Sel.Name
}

func isRecoverCall(
	expression ast.Expr,
) (isRecover bool) {
	call, isCall := expression.(*ast.CallExpr)
	return isCall && len(call.Args) == 0 && isBuiltin(
		call.Fun,
		recoverName,
	)
}

func isBuiltin(
	expression ast.Expr,
	name string,
) (builtin bool) {
	identifier, isIdent := expression.(*ast.Ident)
	return isIdent && identifier.Obj == nil && identifier.Name == name
}

func ignoredFindings(
	file source.File,
) (findings []report.Finding) {
	isSkipped := source.IsTest(file.Path) || source.IsMock(file.Rel) || file.IsGenerated
	if isSkipped {
		return nil
	}

	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			signatureFindings(
				file,
				function.Name.Name,
				function.Type,
				function.Body,
			)...,
		)
	}

	for _, literal := range source.Nodes[*ast.FuncLit](file.Syntax) {
		findings = append(
			findings,
			signatureFindings(
				file,
				literalName,
				literal.Type,
				literal.Body,
			)...,
		)
	}

	return findings
}

func signatureFindings(
	file source.File,
	name string,
	signature *ast.FuncType,
	body *ast.BlockStmt,
) (findings []report.Finding) {
	if body == nil {
		return nil
	}

	for _, parameter := range signature.Params.List {
		findings = appendIgnored(
			findings,
			file,
			name,
			parameter,
			body,
		)
	}

	return findings
}

func appendIgnored(
	findings []report.Finding,
	file source.File,
	name string,
	parameter *ast.Field,
	body *ast.BlockStmt,
) (appended []report.Finding) {
	isIgnored := isErrorParameter(parameter.Type) && (len(parameter.Names) == 0 || slices.ContainsFunc(
		parameter.Names,
		isUnreadIn(body),
	))
	if !isIgnored {
		return findings
	}

	return append(
		findings,
		finding(
			file,
			parameter,
			fmt.Sprintf(
				ignoredFormat,
				name,
			),
		),
	)
}

func isErrorParameter(
	expression ast.Expr,
) (isError bool) {
	variadic, isVariadic := expression.(*ast.Ellipsis)
	if isVariadic {
		return isErrorType(variadic.Elt)
	}

	return isErrorType(expression)
}

func isUnreadIn(
	body *ast.BlockStmt,
) (unread func(*ast.Ident) bool) {
	references := source.Nodes[*ast.Ident](body)
	return func(
		parameter *ast.Ident,
	) (ignored bool) {
		return parameter.Name == blankName || parameter.Obj != nil && !slices.ContainsFunc(
			references,
			isReferenceTo(parameter.Obj),
		)
	}
}

func isReferenceTo(
	object *ast.Object,
) (matches func(*ast.Ident) bool) {
	return func(
		identifier *ast.Ident,
	) (referenced bool) {
		return identifier.Obj == object
	}
}
