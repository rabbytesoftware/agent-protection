package ctxfirstparam

import (
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	storedFormat = "%s stores the context.Context %s: pass ctx as the first parameter of each call instead of keeping it"
	valueFormat  = "%s.%s takes a method value of the context.Context %s and keeps it: pass ctx as the first parameter and call %s.%s() where it is needed"
	literalPlace = "composite literal"
	assignPlace  = "assignment"
	sendPlace    = "channel send"
)

func storedFindings(
	file source.File,
) (findings []report.Finding) {
	for _, literal := range source.Nodes[*ast.CompositeLit](file.Syntax) {
		findings = slices.Concat(
			findings,
			literalFindings(
				file,
				literal,
			),
		)
	}

	for _, assignment := range source.Nodes[*ast.AssignStmt](file.Syntax) {
		findings = slices.Concat(
			findings,
			assignmentFindings(
				file,
				assignment,
			),
		)
	}

	for _, statement := range source.Nodes[*ast.SendStmt](file.Syntax) {
		findings = appendStored(
			findings,
			file,
			sendPlace,
			statement.Value,
		)
	}

	return append(
		findings,
		methodValueFindings(file)...,
	)
}

func literalFindings(
	file source.File,
	literal *ast.CompositeLit,
) (findings []report.Finding) {
	for _, element := range literal.Elts {
		findings = appendStored(
			findings,
			file,
			literalPlace,
			elementValue(element),
		)
	}

	return findings
}

func elementValue(
	element ast.Expr,
) (stored ast.Expr) {
	pair, isPair := element.(*ast.KeyValueExpr)
	if isPair {
		return pair.Value
	}

	return element
}

func assignmentFindings(
	file source.File,
	assignment *ast.AssignStmt,
) (findings []report.Finding) {
	if len(assignment.Lhs) != len(assignment.Rhs) {
		return nil
	}

	for index, target := range assignment.Lhs {
		findings = appendField(
			findings,
			file,
			target,
			assignment.Rhs[index],
		)
	}

	return findings
}

func appendField(
	findings []report.Finding,
	file source.File,
	target ast.Expr,
	value ast.Expr,
) (appended []report.Finding) {
	_, isPlain := ast.Unparen(target).(*ast.Ident)
	if isPlain {
		return findings
	}

	return appendStored(
		findings,
		file,
		assignPlace,
		value,
	)
}

func appendStored(
	findings []report.Finding,
	file source.File,
	place string,
	value ast.Expr,
) (appended []report.Finding) {
	name, isIdent := ast.Unparen(value).(*ast.Ident)
	isContext := isIdent && isContextValue(name)
	if !isContext {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(value),
			Snippet: fmt.Sprintf(
				storedFormat,
				place,
				name.Name,
			),
		},
	)
}

func methodValueFindings(
	file source.File,
) (findings []report.Finding) {
	called := []ast.Expr{}
	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		called = append(
			called,
			ast.Unparen(call.Fun),
		)
	}

	for _, selector := range source.Nodes[*ast.SelectorExpr](file.Syntax) {
		findings = appendMethodValue(
			findings,
			file,
			selector,
			called,
		)
	}

	return findings
}

func appendMethodValue(
	findings []report.Finding,
	file source.File,
	selector *ast.SelectorExpr,
	called []ast.Expr,
) (appended []report.Finding) {
	name, isIdent := ast.Unparen(selector.X).(*ast.Ident)
	isValue := isIdent && isContextValue(name) && !slices.Contains(
		called,
		ast.Expr(selector),
	)
	if !isValue {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(selector),
			Snippet: fmt.Sprintf(
				valueFormat,
				name.Name,
				selector.Sel.Name,
				name.Name,
				name.Name,
				selector.Sel.Name,
			),
		},
	)
}

func isContextValue(
	name *ast.Ident,
) (isContext bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.Field:
		return !isNotContext(declaration)
	case *ast.AssignStmt:
		return isDerivedContext(
			name,
			declaration,
		)
	}

	return false
}

func isDerivedContext(
	name *ast.Ident,
	assignment *ast.AssignStmt,
) (derived bool) {
	isSingleCall := len(assignment.Rhs) == 1 && len(assignment.Lhs) > 0
	if !isSingleCall {
		return false
	}

	first, isIdent := assignment.Lhs[0].(*ast.Ident)
	isFirst := isIdent && first.Name == name.Name
	return isFirst && isDerivingCall(assignment.Rhs[0])
}

func isDerivingCall(
	value ast.Expr,
) (deriving bool) {
	call, isCall := value.(*ast.CallExpr)
	if !isCall {
		return false
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isQualifier := selector.X.(*ast.Ident)
	return isQualifier && qualifier.Name == contextPackage && slices.Contains(
		derivingFunctions(),
		selector.Sel.Name,
	)
}

func derivingFunctions() (names []string) {
	return []string{
		"Background",
		"TODO",
		"WithCancel",
		"WithCancelCause",
		"WithDeadline",
		"WithDeadlineCause",
		"WithTimeout",
		"WithTimeoutCause",
		"WithValue",
		"WithoutCancel",
	}
}
