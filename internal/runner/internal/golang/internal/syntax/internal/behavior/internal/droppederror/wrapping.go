package droppederror

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	formatPackage   = "fmt"
	errorfName      = "Errorf"
	errorMethod     = "Error"
	sentinelPrefix  = "Err"
	wrapVerb        = 'w'
	verbMarker      = "%"
	escapedMarker   = "%%"
	directiveFlags  = "+-# 0123456789."
	unplainPrefixes = "[*"
	wrappingFormat  = "%s is formatted with %%%c: wrap the error with %%w so callers can errors.Is/As it"
	rebuiltFormat   = "%s rebuilds the error from its text and drops the chain: wrap it with fmt.Errorf(\"...: %%w\", %s) so callers can errors.Is/As it"
	newName         = "New"
)

func wrappingFindings(
	file source.File,
	callees map[string]int,
) (findings []report.Finding) {
	calls := slices.DeleteFunc(
		source.Nodes[*ast.CallExpr](file.Syntax),
		isNotErrorf,
	)
	for _, call := range calls {
		findings = append(
			findings,
			errorfFindings(
				file,
				call,
				callees,
			)...,
		)
	}

	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		findings = append(
			findings,
			rebuiltFindings(
				file,
				call,
				callees,
			)...,
		)
	}

	return findings
}

func rebuiltFindings(
	file source.File,
	call *ast.CallExpr,
	callees map[string]int,
) (findings []report.Finding) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	isNew := isSelector &&
		len(call.Args) == 1 &&
		identifierName(selector.X) == errorsPackage &&
		selector.Sel.Name == newName
	if !isNew {
		return nil
	}

	for _, part := range source.Nodes[ast.Expr](call.Args[0]) {
		findings = appendRebuilt(
			findings,
			file,
			call,
			part,
			callees,
		)
	}

	return findings[:min(
		len(findings),
		1,
	)]
}

func appendRebuilt(
	findings []report.Finding,
	file source.File,
	call *ast.CallExpr,
	part ast.Expr,
	callees map[string]int,
) (extended []report.Finding) {
	if !isErrorArgument(
		part,
		callees,
	) {
		return findings
	}

	return append(
		findings,
		finding(
			file,
			call,
			fmt.Sprintf(
				rebuiltFormat,
				types.ExprString(call),
				types.ExprString(errorValue(part)),
			),
		),
	)
}

func errorValue(
	part ast.Expr,
) (value ast.Expr) {
	call, isCall := ast.Unparen(part).(*ast.CallExpr)
	if !isCall {
		return part
	}

	return receiverOf(call)
}

func errorfFindings(
	file source.File,
	call *ast.CallExpr,
	callees map[string]int,
) (findings []report.Finding) {
	verbs := formatVerbs(call.Args[0])
	arguments := call.Args[1:]
	paired := min(
		len(verbs),
		len(arguments),
	)
	for index, verb := range verbs[:paired] {
		findings = appendUnwrapped(
			findings,
			file,
			arguments[index],
			verb,
			callees,
		)
	}

	return findings
}

func appendUnwrapped(
	findings []report.Finding,
	file source.File,
	argument ast.Expr,
	verb rune,
	callees map[string]int,
) (extended []report.Finding) {
	isUnwrapped := verb != wrapVerb && isErrorArgument(
		argument,
		callees,
	)
	if !isUnwrapped {
		return findings
	}

	return append(
		findings,
		finding(
			file,
			argument,
			fmt.Sprintf(
				wrappingFormat,
				types.ExprString(argument),
				verb,
			),
		),
	)
}

func isNotErrorf(
	call *ast.CallExpr,
) (excluded bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	isErrorf := isSelector &&
		len(call.Args) > 1 &&
		identifierName(selector.X) == formatPackage &&
		selector.Sel.Name == errorfName
	return !isErrorf
}

func formatVerbs(
	format ast.Expr,
) (verbs []rune) {
	literal, isLiteral := format.(*ast.BasicLit)
	isString := isLiteral && literal.Kind == token.STRING
	if !isString {
		return nil
	}

	text := constant.StringVal(constant.MakeFromLiteral(
		literal.Value,
		token.STRING,
		0,
	))
	directives := strings.Split(
		strings.ReplaceAll(
			text,
			escapedMarker,
			"",
		),
		verbMarker,
	)
	for _, directive := range directives[1:] {
		verbs = append(
			verbs,
			directiveVerb(directive),
		)
	}

	if slices.Contains(
		verbs,
		utf8.RuneError,
	) {
		return nil
	}

	return verbs
}

func directiveVerb(
	directive string,
) (verb rune) {
	rest := strings.TrimLeft(
		directive,
		directiveFlags,
	)
	verb, _ = utf8.DecodeRuneInString(rest)
	if strings.ContainsRune(
		unplainPrefixes,
		verb,
	) {
		return utf8.RuneError
	}

	return verb
}

func isErrorArgument(
	argument ast.Expr,
	callees map[string]int,
) (isError bool) {
	switch typed := ast.Unparen(argument).(type) {
	case *ast.Ident:
		return isSentinel(typed.Name) || isCaughtError(
			typed,
			callees,
		)
	case *ast.SelectorExpr:
		return isSentinel(typed.Sel.Name)
	case *ast.CallExpr:
		return isErrorMethodCall(typed)
	}

	return false
}

func isErrorMethodCall(
	call *ast.CallExpr,
) (isErrorCall bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	return isSelector && len(call.Args) == 0 && selector.Sel.Name == errorMethod
}

func isSentinel(
	name string,
) (sentinel bool) {
	rest, hasPrefix := strings.CutPrefix(
		name,
		sentinelPrefix,
	)
	first, _ := utf8.DecodeRuneInString(rest)
	return hasPrefix && unicode.IsUpper(first)
}
