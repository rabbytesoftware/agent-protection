package realinfra

import "go/ast"

type binding struct {
	names    []string
	value    ast.Expr
	declared ast.Expr
}
