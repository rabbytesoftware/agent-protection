package cqs

import (
	"go/ast"
	"slices"
)

const (
	receiverPosition = -1
	blankName        = "_"
	methodSeparator  = "."
)

type callable struct {
	key         string
	declaration *ast.FuncDecl
	pointers    map[string]int
	types       map[string]string
	calls       []*ast.CallExpr
}

func callableOf(
	declaration *ast.FuncDecl,
) (described callable) {
	return callable{
		key:         declarationKey(declaration),
		declaration: declaration,
		pointers:    pointerPositions(declaration),
		types:       namedTypes(declaration),
		calls:       liveNodes[*ast.CallExpr](declaration.Body),
	}
}

func declarationKey(
	declaration *ast.FuncDecl,
) (key string) {
	receivers := flatParameters(declaration.Recv)
	if len(receivers) == 0 {
		return declaration.Name.Name
	}

	return methodKey(
		typeName(receivers[0].Type),
		declaration.Name.Name,
	)
}

func methodKey(
	receiverType string,
	method string,
) (key string) {
	return receiverType + methodSeparator + method
}

func pointerPositions(
	declaration *ast.FuncDecl,
) (positions map[string]int) {
	positions = map[string]int{}
	for _, name := range pointerSlots(flatParameters(declaration.Recv)) {
		positions[name] = receiverPosition
	}

	for position, name := range pointerSlots(flatParameters(declaration.Type.Params)) {
		positions[name] = position
	}

	delete(
		positions,
		"",
	)
	return positions
}

func namedTypes(
	declaration *ast.FuncDecl,
) (types map[string]string) {
	types = map[string]string{}
	for _, parameter := range slices.Concat(
		flatParameters(declaration.Recv),
		flatParameters(declaration.Type.Params),
	) {
		types[parameter.Names[0].Name] = typeName(parameter.Type)
	}

	return types
}

func flatParameters(
	list *ast.FieldList,
) (parameters []*ast.Field) {
	if list == nil {
		return nil
	}

	for _, field := range list.List {
		parameters = append(
			parameters,
			splitField(field)...,
		)
	}

	return parameters
}

func splitField(
	field *ast.Field,
) (parameters []*ast.Field) {
	names := field.Names
	if len(names) == 0 {
		names = []*ast.Ident{ast.NewIdent(blankName)}
	}

	for _, name := range names {
		parameters = append(
			parameters,
			&ast.Field{Names: []*ast.Ident{name}, Type: field.Type},
		)
	}

	return parameters
}

func pointerSlots(
	parameters []*ast.Field,
) (names []string) {
	for _, parameter := range parameters {
		names = append(
			names,
			pointerName(parameter),
		)
	}

	return names
}

func pointerName(
	parameter *ast.Field,
) (name string) {
	_, isPointer := parameter.Type.(*ast.StarExpr)
	if !isPointer {
		return ""
	}

	return parameter.Names[0].Name
}

func typeName(
	expression ast.Expr,
) (name string) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return typeName(typed.X)
	case *ast.IndexExpr:
		return typeName(typed.X)
	case *ast.IndexListExpr:
		return typeName(typed.X)
	}

	return ""
}
