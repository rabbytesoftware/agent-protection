package entrypoint

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const anyName = "any"

func compositeVerb(
	context delegation,
	literal *ast.CompositeLit,
) (verb string) {
	if !isBehaviorTable(
		context.unit,
		literal.Type,
	) {
		return noLogic
	}

	return "builds a table of funcs or interfaces"
}

func isBehaviorTable(
	found unit,
	expression ast.Expr,
) (isTable bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.ArrayType:
		return isBehaviorType(
			found,
			typed.Elt,
		)
	case *ast.MapType:
		return isBehaviorType(
			found,
			typed.Value,
		)
	}

	return false
}

func isBehaviorType(
	found unit,
	expression ast.Expr,
) (isBehavior bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.FuncType, *ast.InterfaceType:
		return true
	case *ast.Ident:
		return typed.Name == anyName || slices.Contains(
			found.interfaces,
			typed.Name,
		)
	}

	return false
}

func indexVerb(
	context delegation,
	expression *ast.IndexExpr,
) (verb string) {
	if slices.Contains(
		context.typeIndexes,
		ast.Node(expression),
	) {
		return noLogic
	}

	return "indexes"
}

func typeIndexes(
	found unit,
	body *ast.BlockStmt,
) (indexes []ast.Node) {
	for node := range ast.Preorder(body) {
		indexes = append(
			indexes,
			nodeTypeIndexes(
				found,
				node,
			)...,
		)
	}

	return indexes
}

func nodeTypeIndexes(
	found unit,
	node ast.Node,
) (indexes []ast.Node) {
	switch typed := node.(type) {
	case *ast.CompositeLit:
		return indexesIn(typed.Type)
	case *ast.ValueSpec:
		return indexesIn(typed.Type)
	case *ast.TypeAssertExpr:
		return indexesIn(typed.Type)
	case *ast.CallExpr:
		return genericIndex(
			found,
			typed.Fun,
		)
	}

	return nil
}

func indexesIn(
	expression ast.Expr,
) (indexes []ast.Node) {
	if expression == nil {
		return nil
	}

	for _, index := range source.Nodes[*ast.IndexExpr](expression) {
		indexes = append(
			indexes,
			index,
		)
	}

	return indexes
}

func genericIndex(
	found unit,
	function ast.Expr,
) (indexes []ast.Node) {
	index, isIndex := ast.Unparen(function).(*ast.IndexExpr)
	isGeneric := isIndex && isDeclaredCallee(
		found,
		index.X,
	)
	if !isGeneric {
		return nil
	}

	return []ast.Node{index}
}

func isDeclaredCallee(
	found unit,
	expression ast.Expr,
) (isDeclared bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return slices.Contains(
			found.names,
			typed.Name,
		)
	case *ast.SelectorExpr:
		return isPackageQualifier(
			found,
			typed.X,
		)
	}

	return false
}

func isPackageQualifier(
	found unit,
	expression ast.Expr,
) (isQualifier bool) {
	qualifier, isIdentifier := expression.(*ast.Ident)
	return isIdentifier && len(importsNamed(
		found.file.Syntax,
		qualifier.Name,
	)) > 0
}
