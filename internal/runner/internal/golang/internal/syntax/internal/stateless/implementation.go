package stateless

import (
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const methodFormat = "%s/%d/%d"

func interfaceMethods(
	files []source.File,
) (keys []string) {
	keys = standardMethods()
	for _, file := range files {
		keys = append(
			keys,
			fileInterfaceMethods(file)...,
		)
	}

	return keys
}

func fileInterfaceMethods(
	file source.File,
) (keys []string) {
	for _, declared := range source.Nodes[*ast.InterfaceType](file.Syntax) {
		keys = append(
			keys,
			declaredMethods(declared)...,
		)
	}

	return keys
}

func declaredMethods(
	declared *ast.InterfaceType,
) (keys []string) {
	for _, field := range declared.Methods.List {
		keys = append(
			keys,
			fieldMethods(field)...,
		)
	}

	return keys
}

func fieldMethods(
	field *ast.Field,
) (keys []string) {
	function, isFunction := field.Type.(*ast.FuncType)
	if !isFunction {
		return nil
	}

	for _, name := range field.Names {
		keys = append(
			keys,
			methodKey(
				name.Name,
				function,
			),
		)
	}

	return keys
}

func methodKey(
	name string,
	function *ast.FuncType,
) (key string) {
	return fmt.Sprintf(
		methodFormat,
		name,
		function.Params.NumFields(),
		function.Results.NumFields(),
	)
}

func implementsOnly(
	files []source.File,
	name string,
	known []string,
) (implements bool) {
	methods := typeMethods(
		files,
		name,
	)
	implements = len(methods) > 0
	for _, method := range methods {
		implements = implements && slices.Contains(
			known,
			methodKey(
				method.Name.Name,
				method.Type,
			),
		)
	}

	return implements
}

func typeMethods(
	files []source.File,
	name string,
) (methods []*ast.FuncDecl) {
	for _, file := range files {
		methods = append(
			methods,
			fileMethods(
				file,
				name,
			)...,
		)
	}

	return methods
}

func fileMethods(
	file source.File,
	name string,
) (methods []*ast.FuncDecl) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		methods = appendMethod(
			methods,
			function,
			name,
		)
	}

	return methods
}

func appendMethod(
	methods []*ast.FuncDecl,
	function *ast.FuncDecl,
	name string,
) (extended []*ast.FuncDecl) {
	isOwn := !isNotMethod(function) && typeName(function.Recv.List[0].Type) == name
	if !isOwn {
		return methods
	}

	return append(
		methods,
		function,
	)
}

func standardMethods() (keys []string) {
	return []string{
		"Write/1/2",
		"Read/1/2",
		"Close/0/1",
		"Len/0/1",
		"Less/2/1",
		"Swap/2/0",
		"ServeHTTP/2/0",
		"Error/0/1",
		"String/0/1",
		"MarshalJSON/0/2",
		"UnmarshalJSON/1/1",
	}
}
