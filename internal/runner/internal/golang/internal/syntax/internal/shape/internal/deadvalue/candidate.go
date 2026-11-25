package deadvalue

import "go/ast"

type candidate struct {
	kind       string
	name       *ast.Ident
	value      ast.Expr
	referenced bool
	directory  string
}
