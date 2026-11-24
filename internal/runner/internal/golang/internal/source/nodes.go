package source

import "go/ast"

func Nodes[T ast.Node](
	root ast.Node,
) (nodes []T) {
	for node := range ast.Preorder(root) {
		nodes = appendMatching(
			nodes,
			node,
		)
	}

	return nodes
}

func appendMatching[T ast.Node](
	nodes []T,
	node ast.Node,
) (matching []T) {
	typed, matches := node.(T)
	if !matches {
		return nodes
	}

	return append(
		nodes,
		typed,
	)
}
