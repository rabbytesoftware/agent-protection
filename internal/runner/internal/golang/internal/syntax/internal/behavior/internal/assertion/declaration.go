package assertion

import (
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

type declaration struct {
	file     source.File
	imports  map[string]string
	name     string
	receiver *ast.FieldList
	function *ast.FuncType
	body     *ast.BlockStmt
}
