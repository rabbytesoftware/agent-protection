package deadstore

import "go/ast"

type frame struct {
	loop      ast.Node
	breaks    []reaching
	continues []reaching
	carried   reaching
}
