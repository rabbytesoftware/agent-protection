package stringcomment

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	escapeLogSnippet   = "%s message %q is a note about provisional code (a TODO marker or an escape-hatch word): a log line says what happened, not what is left to fix; delete the note and fix the code, and track the work outside the source"
	sentenceLogSnippet = "%s message %q is a sentence explaining the test: a comment moved into a log line; delete it and let the test name and assertions say it"
	logPrefix          = "Log"
	excerptRunes       = 48
	excerptEllipsis    = "..."
)

func logFindings(
	file source.File,
) (findings []report.Finding) {
	calls := logCalls(file)
	findings = proseFindings(
		file,
		calls,
		escapeLogSnippet,
		isEscapeProse,
	)
	if !source.IsTest(file.Path) {
		return findings
	}

	return append(
		findings,
		proseFindings(
			file,
			slices.DeleteFunc(
				slices.Clone(calls),
				isNotTestLog,
			),
			sentenceLogSnippet,
			isPlainSentence,
		)...,
	)
}

func logCalls(
	file source.File,
) (calls []*ast.CallExpr) {
	return slices.DeleteFunc(
		source.Nodes[*ast.CallExpr](file.Syntax),
		isNotLogCall,
	)
}

func proseFindings(
	file source.File,
	calls []*ast.CallExpr,
	format string,
	matches func(string) bool,
) (findings []report.Finding) {
	for _, call := range calls {
		findings = append(
			findings,
			callFindings(
				file,
				call,
				format,
				matches,
			)...,
		)
	}

	return findings
}

func callFindings(
	file source.File,
	call *ast.CallExpr,
	format string,
	matches func(string) bool,
) (findings []report.Finding) {
	for _, argument := range call.Args {
		findings = append(
			findings,
			argumentFindings(
				file,
				types.ExprString(call.Fun),
				argument,
				format,
				matches,
			)...,
		)
	}

	return findings
}

func argumentFindings(
	file source.File,
	callee string,
	argument ast.Expr,
	format string,
	matches func(string) bool,
) (findings []report.Finding) {
	text, isText := stringText(argument)
	if !isText ||
		!matches(text) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(argument),
		Snippet: fmt.Sprintf(
			format,
			callee,
			excerpt(text),
		),
	}}
}

func isPlainSentence(
	text string,
) (sentence bool) {
	return !isEscapeProse(text) && isSentence(text)
}

func isNotTestLog(
	call *ast.CallExpr,
) (excluded bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	return !isSelector || !slices.Contains(
		testLogMethods(),
		selector.Sel.Name,
	)
}

func isNotLogCall(
	call *ast.CallExpr,
) (excluded bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	return !isSelector || !isLogCall(selector)
}

func isLogCall(
	selector *ast.SelectorExpr,
) (logging bool) {
	method := selector.Sel.Name
	isLogPackage := slices.Contains(
		logPackages(),
		qualifierName(selector.X),
	)
	return isLogPackage ||
		strings.HasPrefix(
			method,
			logPrefix,
		) ||
		slices.Contains(
			logMethods(),
			method,
		)
}

func qualifierName(
	expression ast.Expr,
) (name string) {
	ident, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return ident.Name
}

func excerpt(
	text string,
) (shortened string) {
	runes := []rune(text)
	if len(runes) <= excerptRunes {
		return text
	}

	return string(runes[:excerptRunes]) + excerptEllipsis
}

func logPackages() (names []string) {
	return []string{
		"log",
		"slog",
	}
}

func testLogMethods() (names []string) {
	return []string{
		"Log",
		"Logf",
	}
}

func logMethods() (names []string) {
	return []string{
		"Debug",
		"Debugf",
		"Debugw",
		"Debugln",
		"DebugContext",
		"Info",
		"Infof",
		"Infow",
		"Infoln",
		"InfoContext",
		"Warn",
		"Warnf",
		"Warnw",
		"Warnln",
		"WarnContext",
		"Warning",
		"Warningf",
		"Skip",
		"Skipf",
	}
}
