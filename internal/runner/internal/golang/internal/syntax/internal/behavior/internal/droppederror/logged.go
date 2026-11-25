package droppederror

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	logPackage        = "log"
	structuredPackage = "slog"
	osPackage         = "os"
	exitName          = "Exit"
	panicName         = "panic"
	fatalPrefix       = "Fatal"
	panicPrefix       = "Panic"
	loggedFormat      = "%s(...) logs %s and carries on: logging an error and carrying on swallows it: return it"
	printPrefix       = "Print"
	logPrefix         = "Log"
	streamPrefix      = "Fprint"
	contextSuffix     = "Context"
	formatSuffix      = "f"
	stderrName        = "Stderr"
)

func loggedFindings(
	file source.File,
	function *ast.FuncDecl,
	callees map[string]int,
) (findings []report.Finding) {
	blocks := slices.DeleteFunc(
		source.Nodes[*ast.BlockStmt](function),
		isTerminated,
	)
	for _, block := range blocks {
		findings = append(
			findings,
			blockLoggedFindings(
				file,
				function,
				block,
				callees,
			)...,
		)
	}

	return findings
}

func blockLoggedFindings(
	file source.File,
	function *ast.FuncDecl,
	block *ast.BlockStmt,
	callees map[string]int,
) (findings []report.Finding) {
	for _, statement := range block.List {
		findings = append(
			findings,
			statementLoggedFindings(
				file,
				function,
				block,
				loggingCall(statement),
				callees,
			)...,
		)
	}

	return findings
}

func statementLoggedFindings(
	file source.File,
	function *ast.FuncDecl,
	block *ast.BlockStmt,
	call *ast.CallExpr,
	callees map[string]int,
) (findings []report.Finding) {
	if call == nil {
		return nil
	}

	for _, logged := range loggedErrors(
		call.Args,
		callees,
	) {
		findings = appendSwallowed(
			findings,
			file,
			function,
			block,
			call,
			logged,
		)
	}

	return findings
}

func appendSwallowed(
	findings []report.Finding,
	file source.File,
	function *ast.FuncDecl,
	block *ast.BlockStmt,
	call *ast.CallExpr,
	logged *ast.Ident,
) (extended []report.Finding) {
	isHandled := isParameter(
		function,
		logged,
	) || isCarriedOn(
		function,
		block,
		logged,
	)
	if isHandled {
		return findings
	}

	return append(
		findings,
		finding(
			file,
			call,
			fmt.Sprintf(
				loggedFormat,
				types.ExprString(call.Fun),
				logged.Name,
			),
		),
	)
}

func isCarriedOn(
	function *ast.FuncDecl,
	block *ast.BlockStmt,
	logged *ast.Ident,
) (carried bool) {
	uses, bindings := occurrences(
		function,
		logged,
	)
	return anyBetween(
		uses,
		block.End(),
		earliestAfter(
			bindings,
			block.End(),
			function.End(),
		),
	)
}

func loggingCall(
	statement ast.Stmt,
) (call *ast.CallExpr) {
	expression, isExpression := statement.(*ast.ExprStmt)
	if !isExpression {
		return nil
	}

	call, isCall := expression.X.(*ast.CallExpr)
	isLoggingCall := isCall && isLogging(call)
	if !isLoggingCall {
		return nil
	}

	return call
}

func isLogging(
	call *ast.CallExpr,
) (logging bool) {
	switch typed := call.Fun.(type) {
	case *ast.Ident:
		return slices.Contains(
			builtinPrints(),
			typed.Name,
		) && typed.Obj == nil
	case *ast.SelectorExpr:
		return isSelectorLogging(
			call,
			typed,
		)
	}

	return false
}

func isSelectorLogging(
	call *ast.CallExpr,
	selector *ast.SelectorExpr,
) (logging bool) {
	receiver, isIdent := selector.X.(*ast.Ident)
	isPackage := isIdent && receiver.Obj == nil && slices.Contains(
		loggingPackages(),
		receiver.Name,
	)
	if isPackage {
		return isPackageLogging(
			receiver.Name,
			selector.Sel.Name,
		) || isStreamPrint(
			receiver.Name,
			call,
		)
	}

	isUnresolved := isIdent && receiver.Obj == nil
	return !isUnresolved && isLoggingMethod(selector.Sel.Name) && !isTestingReceiver(selector.X)
}

func isLoggingMethod(
	name string,
) (logging bool) {
	isPrefixed := strings.HasPrefix(
		name,
		printPrefix,
	) || strings.HasPrefix(
		name,
		logPrefix,
	)
	return isPrefixed || slices.Contains(
		levelMethods(),
		strings.TrimSuffix(
			strings.TrimSuffix(
				name,
				contextSuffix,
			),
			formatSuffix,
		),
	)
}

func isStreamPrint(
	pkg string,
	call *ast.CallExpr,
) (printing bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	isFprint := pkg == formatPackage &&
		isSelector &&
		len(call.Args) > 1 &&
		strings.HasPrefix(
			selector.Sel.Name,
			streamPrefix,
		)
	if !isFprint {
		return false
	}

	stream, isStream := call.Args[0].(*ast.SelectorExpr)
	return isStream && identifierName(stream.X) == osPackage && stream.Sel.Name == stderrName
}

func isTestingReceiver(
	receiver ast.Expr,
) (testing bool) {
	name := ""
	switch typed := ast.Unparen(receiver).(type) {
	case *ast.Ident:
		name = typed.Name
	case *ast.SelectorExpr:
		name = typed.Sel.Name
	}

	return slices.Contains(
		testingNames(),
		name,
	)
}

func isPackageLogging(
	pkg string,
	name string,
) (logging bool) {
	switch pkg {
	case formatPackage:
		return slices.Contains(
			formatPrints(),
			name,
		)
	case logPackage:
		return !isTerminatingLog(name)
	case structuredPackage:
		return true
	}

	return false
}

func isTerminatingLog(
	name string,
) (terminating bool) {
	return strings.HasPrefix(
		name,
		fatalPrefix,
	) || strings.HasPrefix(
		name,
		panicPrefix,
	)
}

func loggedErrors(
	arguments []ast.Expr,
	callees map[string]int,
) (logged []*ast.Ident) {
	for _, argument := range arguments {
		logged = append(
			logged,
			argumentErrors(
				argument,
				callees,
			)...,
		)
	}

	return logged
}

func argumentErrors(
	argument ast.Expr,
	callees map[string]int,
) (logged []*ast.Ident) {
	switch typed := ast.Unparen(argument).(type) {
	case *ast.Ident:
		return failureOnly(
			typed,
			callees,
		)
	case *ast.CallExpr:
		return callErrors(
			typed,
			callees,
		)
	}

	return nil
}

func failureOnly(
	name *ast.Ident,
	callees map[string]int,
) (logged []*ast.Ident) {
	if !isCaughtError(
		name,
		callees,
	) {
		return nil
	}

	return []*ast.Ident{name}
}

func callErrors(
	call *ast.CallExpr,
	callees map[string]int,
) (logged []*ast.Ident) {
	receiver, isIdent := ast.Unparen(receiverOf(call)).(*ast.Ident)
	isIdentError := isIdent && isErrorMethodCall(call)
	if isIdentError {
		return []*ast.Ident{receiver}
	}

	return loggedErrors(
		call.Args,
		callees,
	)
}

func receiverOf(
	call *ast.CallExpr,
) (receiver ast.Expr) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}

	return selector.X
}

func isTerminated(
	block *ast.BlockStmt,
) (terminated bool) {
	return slices.ContainsFunc(
		block.List,
		isTerminating,
	)
}

func isTerminating(
	statement ast.Stmt,
) (terminating bool) {
	switch typed := statement.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return typed.Tok == token.GOTO
	case *ast.ExprStmt:
		return isTerminatingCall(typed.X)
	}

	return false
}

func isTerminatingCall(
	expression ast.Expr,
) (terminating bool) {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return false
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return isBuiltin(
			call.Fun,
			panicName,
		)
	}

	pkg := identifierName(selector.X)
	isExit := pkg == osPackage && selector.Sel.Name == exitName
	return isExit || pkg == logPackage && isTerminatingLog(selector.Sel.Name)
}

func loggingPackages() (names []string) {
	return []string{
		formatPackage,
		logPackage,
		structuredPackage,
	}
}

func testingNames() (names []string) {
	return []string{
		"t",
		"b",
		"f",
		"tb",
	}
}

func levelMethods() (names []string) {
	return []string{
		"Error",
		"Warn",
		"Warning",
		"Info",
		"Debug",
		"Trace",
	}
}

func builtinPrints() (names []string) {
	return []string{"print", "println"}
}

func formatPrints() (names []string) {
	return []string{"Print", "Printf", "Println"}
}
