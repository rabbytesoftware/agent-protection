package cqs

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func liveNodes[T ast.Node](
	body *ast.BlockStmt,
) (nodes []T) {
	dormant := dormantLiterals(body)
	for _, node := range source.Nodes[T](body) {
		nodes = appendLive(
			nodes,
			node,
			dormant,
		)
	}

	return nodes
}

func appendLive[T ast.Node](
	nodes []T,
	node T,
	dormant []*ast.FuncLit,
) (extended []T) {
	inside := false
	for _, literal := range dormant {
		inside = inside || encloses(
			literal,
			node,
		)
	}

	if inside {
		return nodes
	}

	return append(
		nodes,
		node,
	)
}

func encloses(
	literal *ast.FuncLit,
	node ast.Node,
) (enclosed bool) {
	return literal.Pos() <= node.Pos() && node.End() <= literal.End()
}

func dormantLiterals(
	body *ast.BlockStmt,
) (dormant []*ast.FuncLit) {
	running := runningLiterals(body)
	for _, literal := range source.Nodes[*ast.FuncLit](body) {
		dormant = appendDormant(
			dormant,
			literal,
			running,
		)
	}

	return dormant
}

func appendDormant(
	dormant []*ast.FuncLit,
	literal *ast.FuncLit,
	running []*ast.FuncLit,
) (extended []*ast.FuncLit) {
	if slices.Contains(
		running,
		literal,
	) {
		return dormant
	}

	return append(
		dormant,
		literal,
	)
}

func runningLiterals(
	body *ast.BlockStmt,
) (running []*ast.FuncLit) {
	invoked := []ast.Expr{}
	for _, call := range source.Nodes[*ast.CallExpr](body) {
		invoked = append(
			slices.Concat(
				invoked,
				call.Args,
			),
			call.Fun,
		)
	}

	for _, expression := range invoked {
		running = appendLiteral(
			running,
			expression,
		)
	}

	return running
}

func appendLiteral(
	literals []*ast.FuncLit,
	expression ast.Expr,
) (extended []*ast.FuncLit) {
	literal, isLiteral := ast.Unparen(expression).(*ast.FuncLit)
	if !isLiteral {
		return literals
	}

	return append(
		literals,
		literal,
	)
}
