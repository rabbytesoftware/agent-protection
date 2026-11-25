package interfacesize

import "go/ast"

type declaration struct {
	directory string
	spec      *ast.TypeSpec
	body      *ast.InterfaceType
	aliases   map[string]string
}
