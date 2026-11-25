package publicapi

import (
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const errorMethodKind = "error method"

func errorMethodViolations(
	function *ast.FuncDecl,
	receiver string,
) (found []violation) {
	if source.IsPlainRendering(function.Body) {
		return nil
	}

	return []violation{{
		node:     function,
		kind:     errorMethodKind,
		name:     receiver + "." + function.Name.Name,
		exported: true,
	}}
}
