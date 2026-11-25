package deadstore

import (
	"go/ast"
	"go/token"
)

type store struct {
	variable token.Pos
	target   *ast.Ident
	call     string
	loop     ast.Node
}
