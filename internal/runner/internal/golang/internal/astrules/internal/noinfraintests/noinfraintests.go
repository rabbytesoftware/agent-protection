package noinfraintests

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R41"

// Noinfraintests reports every call in a unit test that opens real
// infrastructure (http, net or sql dialing, requesting or opening)
// unless its arguments point at a loopback address, directly or through a
// variable assigned from one. Building in-memory requests is not a call.
func Noinfraintests(
	tests []source.File,
) (findings []report.Finding) {
	for _, file := range tests {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				function,
			)...,
		)
	}

	return findings
}

func functionFindings(
	file source.File,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	loopback := loopbackNames(function.Body)
	calls := slices.DeleteFunc(
		source.Nodes[*ast.CallExpr](function.Body),
		isNotLiveCall,
	)
	for _, call := range calls {
		findings = append(
			findings,
			callFindings(
				file,
				call,
				loopback,
			)...,
		)
	}

	return findings
}

func callFindings(
	file source.File,
	call *ast.CallExpr,
	loopback []string,
) (findings []report.Finding) {
	if reachesLoopback(
		call,
		loopback,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(call),
		Snippet: fmt.Sprintf(
			"unit test makes a live call to %s: inject an in-memory fake",
			types.ExprString(call.Fun),
		),
	}}
}

func isNotLiveCall(
	call *ast.CallExpr,
) (notLive bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return true
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	isLive := isIdent && slices.Contains(
		livePackages(),
		qualifier.Name,
	) && slices.Contains(
		liveCalls(),
		selector.Sel.Name,
	)
	return !isLive
}

func loopbackNames(
	body *ast.BlockStmt,
) (names []string) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](body) {
		names = append(
			names,
			assignedLoopbackNames(assignment)...,
		)
	}

	return names
}

func assignedLoopbackNames(
	assignment *ast.AssignStmt,
) (names []string) {
	for index := range assignment.Rhs {
		names = append(
			names,
			assignedLoopbackName(
				assignment,
				index,
			)...,
		)
	}

	return names
}

func assignedLoopbackName(
	assignment *ast.AssignStmt,
	index int,
) (names []string) {
	isPairedLoopback := index < len(assignment.Lhs) && containsLoopbackHint(assignment.Rhs[index])
	if !isPairedLoopback {
		return nil
	}

	target, isIdent := assignment.Lhs[index].(*ast.Ident)
	if !isIdent {
		return nil
	}

	return []string{target.Name}
}

func reachesLoopback(
	call *ast.CallExpr,
	loopback []string,
) (reaches bool) {
	for _, argument := range call.Args {
		reaches = reaches || containsLoopbackHint(argument) || slices.Contains(
			loopback,
			identName(argument),
		)
	}

	return reaches
}

func identName(
	expression ast.Expr,
) (name string) {
	ident, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return ident.Name
}

func containsLoopbackHint(
	expression ast.Expr,
) (hinted bool) {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		return literalLooksLoopback(typed)
	case *ast.BinaryExpr:
		return containsLoopbackHint(typed.X) || containsLoopbackHint(typed.Y)
	case *ast.CallExpr:
		return slices.ContainsFunc(
			typed.Args,
			containsLoopbackHint,
		) || isLoopbackSelector(typed.Fun)
	case *ast.SelectorExpr:
		return isLoopbackSelector(typed)
	}

	return false
}

func isLoopbackSelector(
	expression ast.Expr,
) (loopback bool) {
	selector, isSelector := expression.(*ast.SelectorExpr)
	return isSelector && slices.Contains(
		loopbackSelectors(),
		selector.Sel.Name,
	)
}

func literalLooksLoopback(
	literal *ast.BasicLit,
) (loopback bool) {
	for _, marker := range loopbackMarkers() {
		loopback = loopback || strings.Contains(
			literal.Value,
			marker,
		)
	}

	return loopback
}

func livePackages() (names []string) {
	return []string{"http", "net", "sql"}
}

func liveCalls() (names []string) {
	return []string{"Get", "Post", "Put", "Patch", "Do", "Dial", "DialContext", "Open", "OpenDB"}
}

func loopbackSelectors() (names []string) {
	return []string{"Addr", "URL", "String"}
}

func loopbackMarkers() (markers []string) {
	return []string{"127.0.0.1", "localhost", "unix"}
}
