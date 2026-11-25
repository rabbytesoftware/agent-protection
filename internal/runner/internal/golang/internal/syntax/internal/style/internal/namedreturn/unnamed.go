package namedreturn

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	unnamedSnippet  = "%s returns unnamed results (%s): name every result after what it holds, for example (%s)"
	literalSubject  = "func literal"
	functionSubject = "func %s"
	methodSubject   = "method %s.%s"
	errorType       = "error"
	errorName       = "err"
	placeholderName = "<name>"
	pluralSuffix    = "s"
	pointerPrefix   = "*"
	listSeparator   = ", "
)

func unnamedFindings(
	file source.File,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			resultFindings(
				file,
				declarationSubject(function),
				function.Type,
				function.Body,
			)...,
		)
	}

	for _, literal := range source.Nodes[*ast.FuncLit](file.Syntax) {
		findings = append(
			findings,
			resultFindings(
				file,
				literalSubject,
				literal.Type,
				literal.Body,
			)...,
		)
	}

	return findings
}

func resultFindings(
	file source.File,
	subject string,
	signature *ast.FuncType,
	body *ast.BlockStmt,
) (findings []report.Finding) {
	isUnnamed := body != nil &&
		signature.Results != nil &&
		len(signature.Results.List) > 0 &&
		!hasNamedResults(signature)
	if !isUnnamed {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(signature.Results),
		Snippet: fmt.Sprintf(
			unnamedSnippet,
			subject,
			resultTypes(signature.Results),
			suggestedResults(signature.Results),
		),
	}}
}

func declarationSubject(
	function *ast.FuncDecl,
) (subject string) {
	isFunction := function.Recv == nil ||
		len(function.Recv.List) == 0
	if isFunction {
		return fmt.Sprintf(
			functionSubject,
			function.Name.Name,
		)
	}

	return fmt.Sprintf(
		methodSubject,
		strings.TrimPrefix(
			types.ExprString(function.Recv.List[0].Type),
			pointerPrefix,
		),
		function.Name.Name,
	)
}

func resultTypes(
	results *ast.FieldList,
) (listed string) {
	written := []string{}
	for _, field := range results.List {
		written = append(
			written,
			types.ExprString(field.Type),
		)
	}

	return strings.Join(
		written,
		listSeparator,
	)
}

func suggestedResults(
	results *ast.FieldList,
) (listed string) {
	written := []string{}
	for _, field := range results.List {
		written = append(
			written,
			suggestedName(field.Type)+" "+types.ExprString(field.Type),
		)
	}

	return strings.Join(
		written,
		listSeparator,
	)
}

func suggestedName(
	kind ast.Expr,
) (name string) {
	written := types.ExprString(kind)
	if written == errorType {
		return errorName
	}

	name = validName(typeName(kind))
	if isLetter(name) {
		return opaqueName
	}

	if isVague(
		name,
		kind,
		nil,
	) {
		return placeholderName
	}

	return name
}

func typeName(
	kind ast.Expr,
) (name string) {
	switch typed := kind.(type) {
	case *ast.Ident:
		return exportedName(typed.Name)
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return typeName(typed.X)
	case *ast.ArrayType:
		return pluralName(typeName(typed.Elt))
	case *ast.IndexExpr:
		return typeName(typed.X)
	case *ast.IndexListExpr:
		return typeName(typed.X)
	}

	return ""
}

func exportedName(
	name string,
) (exported string) {
	if !token.IsExported(name) {
		return ""
	}

	return name
}

func pluralName(
	name string,
) (plural string) {
	if name == "" {
		return ""
	}

	return name + pluralSuffix
}

func validName(
	name string,
) (valid string) {
	runes := []rune(name)
	if len(runes) == 0 {
		return placeholderName
	}

	runes[0] = unicode.ToLower(runes[0])
	lowered := string(runes)
	if token.IsKeyword(lowered) {
		return placeholderName
	}

	return lowered
}
