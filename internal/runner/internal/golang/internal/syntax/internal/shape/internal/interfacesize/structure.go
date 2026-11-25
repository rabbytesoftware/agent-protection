package interfacesize

import "go/ast"

type structure struct {
	owner declaration
	body  *ast.StructType
}
