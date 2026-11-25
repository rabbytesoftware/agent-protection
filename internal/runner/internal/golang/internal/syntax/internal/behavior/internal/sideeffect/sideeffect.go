package sideeffect

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

const rule = "R34"

func Sideeffect(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
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
	functions := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotQuery,
	)
	for _, function := range functions {
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
	writes := writeCalls(function.Body)
	if len(writes) == 0 {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(function),
		Snippet: fmt.Sprintf(
			"func %s(...) calls .%s() despite its query name: rename it to say what it changes or move the write out",
			function.Name.Name,
			writes[0],
		),
	}}
}

func isNotQuery(
	function *ast.FuncDecl,
) (excluded bool) {
	return function.Body == nil || !hasQueryName(function)
}

func hasQueryName(
	function *ast.FuncDecl,
) (isQuery bool) {
	for _, prefix := range queryPrefixes() {
		isQuery = isQuery || startsWithWord(
			function.Name.Name,
			prefix,
		)
	}

	return isQuery
}

func startsWithWord(
	name string,
	word string,
) (starts bool) {
	rest, hasPrefix := strings.CutPrefix(
		name,
		word,
	)
	if !hasPrefix {
		return false
	}

	next, _ := utf8.DecodeRuneInString(rest)
	return rest == "" || unicode.IsUpper(next) || unicode.IsDigit(next)
}

func writeCalls(
	body *ast.BlockStmt,
) (methods []string) {
	for _, call := range source.Nodes[*ast.CallExpr](body) {
		methods = append(
			methods,
			calledMethod(call),
		)
	}

	return slices.DeleteFunc(
		methods,
		isNotWriteVerb,
	)
}

func calledMethod(
	call *ast.CallExpr,
) (method string) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}

	return selector.Sel.Name
}

func isNotWriteVerb(
	method string,
) (excluded bool) {
	return !slices.Contains(
		writeVerbs(),
		method,
	)
}

func queryPrefixes() (prefixes []string) {
	return []string{
		"Get",
		"Find",
		"Compute",
		"List",
		"Calc",
		"Calculate",
		"Is",
		"Has",
	}
}

func writeVerbs() (verbs []string) {
	return []string{
		"Write",
		"WriteFile",
		"Save",
		"Delete",
		"Remove",
		"RemoveAll",
		"Create",
		"Update",
		"Insert",
		"Exec",
		"ExecContext",
		"Post",
		"Put",
		"Patch",
		"Send",
		"Set",
		"MkdirAll",
		"Store",
	}
}
