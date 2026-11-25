package droppederror

import (
	"go/ast"
	"maps"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const errorType = "error"

func erroringCallees(
	files []source.File,
) (callees map[string]map[string]int) {
	callees = map[string]map[string]int{}
	for _, file := range files {
		directory := filepath.Dir(file.Path)
		callees[directory] = merged(
			callees[directory],
			fileCallees(file),
		)
	}

	return callees
}

func merged(
	known map[string]int,
	found map[string]int,
) (union map[string]int) {
	union = map[string]int{}
	maps.Copy(
		union,
		known,
	)
	maps.Copy(
		union,
		found,
	)
	return union
}

func fileCallees(
	file source.File,
) (callees map[string]int) {
	callees = map[string]int{}
	for _, field := range source.Nodes[*ast.Field](file.Syntax) {
		maps.Copy(
			callees,
			fieldCallees(field),
		)
	}

	functions := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotErroringFunction,
	)
	for _, function := range functions {
		callees[function.Name.Name] = resultCount(function.Type)
	}

	return callees
}

func fieldCallees(
	field *ast.Field,
) (callees map[string]int) {
	signature, isFunc := field.Type.(*ast.FuncType)
	isErroring := isFunc && returnsError(signature)
	if !isErroring {
		return nil
	}

	callees = map[string]int{}
	for _, name := range field.Names {
		callees[name.Name] = resultCount(signature)
	}

	return callees
}

func isNotErroringFunction(
	function *ast.FuncDecl,
) (excluded bool) {
	return !returnsError(function.Type)
}

func returnsError(
	signature *ast.FuncType,
) (erroring bool) {
	if signature.Results == nil {
		return false
	}

	results := signature.Results.List
	return isErrorType(results[len(results)-1].Type)
}

func resultCount(
	signature *ast.FuncType,
) (count int) {
	for _, result := range signature.Results.List {
		count += max(
			1,
			len(result.Names),
		)
	}

	return count
}

func isErrorType(
	expression ast.Expr,
) (isError bool) {
	name, isIdent := expression.(*ast.Ident)
	return isIdent && name.Name == errorType
}
