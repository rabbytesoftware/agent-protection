package source

import "go/ast"

// Nodes returns every node of type T inside root, in source order. It is
// the flat alternative to an ast.Inspect closure: callers range over the
// result instead of nesting their logic inside a visitor callback.
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
