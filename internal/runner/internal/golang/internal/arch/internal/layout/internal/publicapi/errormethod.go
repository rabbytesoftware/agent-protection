package publicapi

import (
	"go/ast"
	"go/token"
	"slices"
)

const (
	errorMethodKind = "error method"
	errorsQualifier = "errors"
)

func errorMethodViolations(
	function *ast.FuncDecl,
	receiver string,
) (found []violation) {
	if isPlainRendering(function.Body) {
		return nil
	}

	return []violation{{node: function, kind: errorMethodKind, name: receiver + "." + function.Name.Name, exported: true}}
}

func isPlainRendering(
	body *ast.BlockStmt,
) (isPlain bool) {
	if body == nil {
		return true
	}

	if len(body.List) != 1 {
		return false
	}

	statement, isReturn := body.List[0].(*ast.ReturnStmt)
	return isReturn && len(statement.Results) <= 1 && !slices.ContainsFunc(
		statement.Results,
		isNotPlainResult,
	)
}

func isNotPlainResult(
	expression ast.Expr,
) (excluded bool) {
	call, isCall := ast.Unparen(expression).(*ast.CallExpr)
	isFormatting := isCall && isFormattingCall(call)
	if isFormatting {
		return false
	}

	return !isOperand(expression)
}

func isOperand(
	expression ast.Expr,
) (matches bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.BasicLit, *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return isOperand(typed.X)
	case *ast.BinaryExpr:
		return typed.Op == token.ADD && isOperand(typed.X) && isOperand(typed.Y)
	case *ast.CallExpr:
		return isConversion(typed) || isRenderingCall(typed)
	}

	return false
}

func isConversion(
	call *ast.CallExpr,
) (matches bool) {
	converted, isIdentifier := ast.Unparen(call.Fun).(*ast.Ident)
	return isIdentifier && len(call.Args) == 1 && slices.Contains(
		conversionTypes(),
		converted.Name,
	) && isOperand(call.Args[0])
}

func isRenderingCall(
	call *ast.CallExpr,
) (matches bool) {
	method, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	return isSelector && len(call.Args) == 0 && slices.Contains(
		renderingMethods(),
		method.Sel.Name,
	) && isOperand(method.X)
}

func isFormattingCall(
	call *ast.CallExpr,
) (matches bool) {
	function, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdentifier := function.X.(*ast.Ident)
	isFormatter := isIdentifier && (qualifier.Name == errorsQualifier || slices.Contains(
		formattingFunctions(),
		qualifier.Name+"."+function.Sel.Name,
	))
	return isFormatter && !slices.ContainsFunc(
		call.Args,
		isNotOperand,
	)
}

func isNotOperand(
	expression ast.Expr,
) (excluded bool) {
	return !isOperand(expression)
}

func conversionTypes() (names []string) {
	return []string{"string", "error", "bool", "int", "int64", "uint", "uint64", "float64", "rune", "byte"}
}

func renderingMethods() (names []string) {
	return []string{"Error", "String"}
}

func formattingFunctions() (names []string) {
	return []string{"fmt.Sprintf", "fmt.Errorf", "fmt.Sprint"}
}
