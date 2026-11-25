package packagestate

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R24"
	initFunction   = "init"
	blankName      = "_"
	embedDirective = "//go:embed"
	mainPackage    = "main"
	stringType     = "string"
	initSnippet    = "func init: no init; wire dependencies explicitly in container.go"
	variableFormat = "package variable %s: no mutable package-level state; pass it in through a constructor"
	blankSnippet   = "package variable _ is initialized by a call that runs at import time like init: call it explicitly from container.go or func main; only interface assertions such as var _ I = (*T)(nil) stay"
	nilName        = "nil"
)

func Packagestate(
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
	if source.IsTest(file.Rel) {
		return nil
	}

	for _, decl := range file.Syntax.Decls {
		findings = append(
			findings,
			declFindings(
				file,
				decl,
			)...,
		)
	}

	return append(
		findings,
		slices.DeleteFunc(
			environmentFindings(file),
			isReportedIn(findings),
		)...,
	)
}

func isReportedIn(
	findings []report.Finding,
) (reported func(report.Finding) bool) {
	return func(
		candidate report.Finding,
	) (isReported bool) {
		return slices.ContainsFunc(
			findings,
			isAt(candidate.Location),
		)
	}
}

func isAt(
	location string,
) (matches func(report.Finding) bool) {
	return func(
		finding report.Finding,
	) (isSame bool) {
		return finding.Location == location
	}
}

func declFindings(
	file source.File,
	decl ast.Decl,
) (findings []report.Finding) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return initFindings(
			file,
			typed,
		)
	case *ast.GenDecl:
		return variableFindings(
			file,
			typed,
		)
	}

	return nil
}

func initFindings(
	file source.File,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	isInit := function.Recv == nil && function.Name.Name == initFunction
	if !isInit {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(function),
		Snippet:  initSnippet,
	}}
}

func variableFindings(
	file source.File,
	decl *ast.GenDecl,
) (findings []report.Finding) {
	if decl.Tok != token.VAR {
		return nil
	}

	for _, spec := range decl.Specs {
		findings = append(
			findings,
			specFindings(
				file,
				decl,
				spec,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	decl *ast.GenDecl,
	node ast.Spec,
) (findings []report.Finding) {
	spec, isValue := node.(*ast.ValueSpec)
	isSkipped := !isValue ||
		isEmbed(decl.Doc) ||
		isEmbed(spec.Doc) ||
		isLinkerTarget(
			file,
			spec,
		)
	if isSkipped {
		return nil
	}

	for index, name := range spec.Names {
		findings = append(
			findings,
			nameFindings(
				file,
				name,
				valueAt(
					spec.Values,
					index,
				),
			)...,
		)
	}

	return findings
}

func nameFindings(
	file source.File,
	name *ast.Ident,
	value ast.Expr,
) (findings []report.Finding) {
	if name.Name == blankName {
		return blankFindings(
			file,
			name,
			value,
		)
	}

	isSentinel := isSentinelName(name.Name) && isSentinelValue(value)
	if isSentinel {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(name),
		Snippet: fmt.Sprintf(
			variableFormat,
			name.Name,
		),
	}}
}

func blankFindings(
	file source.File,
	name *ast.Ident,
	value ast.Expr,
) (findings []report.Finding) {
	isInert := value == nil || !slices.ContainsFunc(
		source.Nodes[*ast.CallExpr](value),
		isRunningCall,
	)
	if isInert {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(name),
		Snippet:  blankSnippet,
	}}
}

func isRunningCall(
	call *ast.CallExpr,
) (running bool) {
	conversion, isParen := call.Fun.(*ast.ParenExpr)
	if !isParen {
		return true
	}

	_, isPointer := conversion.X.(*ast.StarExpr)
	return !isPointer || len(call.Args) != 1 || !isNil(call.Args[0])
}

func isNil(
	expression ast.Expr,
) (null bool) {
	identifier, isIdent := expression.(*ast.Ident)
	return isIdent && identifier.Name == nilName
}

func valueAt(
	values []ast.Expr,
	index int,
) (value ast.Expr) {
	isShared := len(values) == 1
	if isShared {
		return values[0]
	}

	if index >= len(values) {
		return nil
	}

	return values[index]
}

func isSentinelName(
	name string,
) (sentinel bool) {
	for _, prefix := range sentinelPrefixes() {
		rest, hasPrefix := strings.CutPrefix(
			name,
			prefix,
		)
		first, _ := utf8.DecodeRuneInString(rest)
		sentinel = sentinel || hasPrefix && unicode.IsUpper(first)
	}

	return sentinel
}

func isSentinelValue(
	value ast.Expr,
) (sentinel bool) {
	switch typed := value.(type) {
	case *ast.CallExpr:
		return len(source.Nodes[*ast.FuncLit](typed)) == 0
	case *ast.SelectorExpr:
		return isSentinelName(typed.Sel.Name)
	}

	return false
}

func isEmbed(
	group *ast.CommentGroup,
) (embedded bool) {
	if group == nil {
		return false
	}

	return slices.ContainsFunc(
		group.List,
		isEmbedComment,
	)
}

func isEmbedComment(
	comment *ast.Comment,
) (embedded bool) {
	return strings.HasPrefix(
		comment.Text,
		embedDirective,
	)
}

func sentinelPrefixes() (prefixes []string) {
	return []string{"Err", "err"}
}

func isLinkerTarget(
	file source.File,
	spec *ast.ValueSpec,
) (target bool) {
	typeName, isNamed := spec.Type.(*ast.Ident)
	return file.Syntax.Name.Name == mainPackage &&
		len(spec.Values) == 0 &&
		isNamed &&
		typeName.Name == stringType
}
