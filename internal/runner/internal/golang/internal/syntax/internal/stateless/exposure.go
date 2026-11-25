package stateless

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func interfaceOnly(
	files []source.File,
	known []string,
) (names []string) {
	for _, name := range emptyStructNames(files) {
		names = appendInterfaceOnly(
			names,
			files,
			name,
			known,
		)
	}

	return names
}

func appendInterfaceOnly(
	names []string,
	files []source.File,
	name string,
	known []string,
) (extended []string) {
	isInterfaceOnly := !ast.IsExported(name) && !isInSignature(
		files,
		name,
	) && implementsOnly(
		files,
		name,
		known,
	)
	if !isInterfaceOnly {
		return names
	}

	return append(
		names,
		name,
	)
}

func isInSignature(
	files []source.File,
	name string,
) (referenced bool) {
	for _, file := range files {
		referenced = referenced || fileSignatureNames(
			file,
			name,
		)
	}

	return referenced
}

func fileSignatureNames(
	file source.File,
	name string,
) (referenced bool) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		referenced = referenced || signatureNames(
			function,
			name,
		)
	}

	return referenced
}

func signatureNames(
	function *ast.FuncDecl,
	name string,
) (names bool) {
	for _, field := range slices.Concat(
		function.Type.Params.List,
		fieldList(function.Type.Results),
	) {
		names = names || typeName(field.Type) == name
	}

	return names
}

func fieldList(
	list *ast.FieldList,
) (fields []*ast.Field) {
	if list == nil {
		return nil
	}

	return list.List
}
