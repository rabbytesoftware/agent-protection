package publicapi

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	funcKind     = "func"
	methodKind   = "method"
	typeKind     = "type"
	blankName    = "_"
	sentinelHead = "Err"
)

type violation struct {
	node     ast.Node
	kind     string
	name     string
	exported bool
}

func violations(
	declaration ast.Decl,
) (found []violation) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return functionViolations(typed)
	case *ast.GenDecl:
		return generalViolations(typed)
	}

	return nil
}

func functionViolations(
	function *ast.FuncDecl,
) (found []violation) {
	name := function.Name.Name
	if function.Recv == nil {
		return nameViolations(
			funcKind,
			[]*ast.Ident{function.Name},
			isBlank,
		)
	}

	receiver := receiverType(function.Recv.List[0].Type)
	isErrorMethod := ast.IsExported(receiver) && slices.Contains(
		source.ErrorMethods(),
		name,
	)
	if isErrorMethod {
		return errorMethodViolations(
			function,
			receiver,
		)
	}

	return []violation{{node: function, kind: methodKind, name: receiver + "." + name, exported: function.Name.IsExported()}}
}

func generalViolations(
	general *ast.GenDecl,
) (found []violation) {
	for _, spec := range general.Specs {
		found = append(
			found,
			specViolations(
				general.Tok,
				spec,
			)...,
		)
	}

	return found
}

func specViolations(
	keyword token.Token,
	spec ast.Spec,
) (found []violation) {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		return nameViolations(
			typeKind,
			[]*ast.Ident{typed.Name},
			isExportedOrBlank,
		)
	case *ast.ValueSpec:
		return nameViolations(
			keyword.String(),
			typed.Names,
			allowedValue(keyword),
		)
	}

	return nil
}

func nameViolations(
	kind string,
	names []*ast.Ident,
	isAllowed func(string) bool,
) (found []violation) {
	for _, name := range names {
		found = append(
			found,
			nameViolation(
				kind,
				name,
				isAllowed,
			)...,
		)
	}

	return found
}

func nameViolation(
	kind string,
	name *ast.Ident,
	isAllowed func(string) bool,
) (found []violation) {
	if isAllowed(name.Name) {
		return nil
	}

	return []violation{{node: name, kind: kind, name: name.Name, exported: name.IsExported()}}
}

func allowedValue(
	keyword token.Token,
) (isAllowed func(string) bool) {
	if keyword == token.VAR {
		return isSentinelOrBlank
	}

	return isExportedOrBlank
}

func receiverType(
	expression ast.Expr,
) (name string) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverType(typed.X)
	case *ast.IndexExpr:
		return receiverType(typed.X)
	case *ast.IndexListExpr:
		return receiverType(typed.X)
	}

	return blankName
}

func isBlank(
	name string,
) (allowed bool) {
	return name == blankName
}

func isExportedOrBlank(
	name string,
) (allowed bool) {
	return name == blankName || ast.IsExported(name)
}

func isSentinelOrBlank(
	name string,
) (allowed bool) {
	return name == blankName || ast.IsExported(name) && strings.HasPrefix(
		name,
		sentinelHead,
	)
}
