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
	references  map[string]int
	declared    map[string]ast.Expr
	fields      map[string]ast.Expr
	calls       []*ast.CallExpr
}

func callableOf(
	declaration *ast.FuncDecl,
	fields map[string]ast.Expr,
) (described callable) {
	return callable{
		key:         declarationKey(declaration),
		declaration: declaration,
		pointers: slotPositions(
			declaration,
			pointerName,
		),
		references: slotPositions(
			declaration,
			referenceName,
		),
		declared: declaredTypes(declaration),
		fields:   fields,
		calls:    liveNodes[*ast.CallExpr](declaration.Body),
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

func slotPositions(
	declaration *ast.FuncDecl,
	slotName func(*ast.Field) string,
) (positions map[string]int) {
	positions = map[string]int{}
	for _, name := range slotNames(
		flatParameters(declaration.Recv),
		slotName,
	) {
		positions[name] = receiverPosition
	}

	for position, name := range slotNames(
		flatParameters(declaration.Type.Params),
		slotName,
	) {
		positions[name] = position
	}

	delete(
		positions,
		"",
	)
	return positions
}

func declaredTypes(
	declaration *ast.FuncDecl,
) (types map[string]ast.Expr) {
	types = map[string]ast.Expr{}
	for _, parameter := range slices.Concat(
		flatParameters(declaration.Recv),
		flatParameters(declaration.Type.Params),
	) {
		types[parameter.Names[0].Name] = parameter.Type
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
			&ast.Field{
				Names: []*ast.Ident{name},
				Type:  field.Type,
			},
		)
	}

	return parameters
}

func slotNames(
	parameters []*ast.Field,
	slotName func(*ast.Field) string,
) (names []string) {
	for _, parameter := range parameters {
		names = append(
			names,
			slotName(parameter),
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

func referenceName(
	parameter *ast.Field,
) (name string) {
	switch typed := parameter.Type.(type) {
	case *ast.StarExpr:
		return ""
	case *ast.ArrayType:
		return sliceName(
			parameter,
			typed,
		)
	}

	return parameter.Names[0].Name
}

func sliceName(
	parameter *ast.Field,
	array *ast.ArrayType,
) (name string) {
	if array.Len != nil {
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

func resolvedType(
	described callable,
	expression ast.Expr,
) (resolved ast.Expr) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return described.declared[typed.Name]
	case *ast.ParenExpr:
		return resolvedType(
			described,
			typed.X,
		)
	case *ast.StarExpr:
		return resolvedType(
			described,
			typed.X,
		)
	case *ast.SelectorExpr:
		return fieldType(
			described.fields,
			resolvedType(
				described,
				typed.X,
			),
			typed.Sel.Name,
		)
	}

	return nil
}

func fieldType(
	fields map[string]ast.Expr,
	owner ast.Expr,
	name string,
) (field ast.Expr) {
	structure, isStruct := elementType(owner).(*ast.StructType)
	if isStruct {
		return structFields(structure)[name]
	}

	return fields[methodKey(
		typeName(owner),
		name,
	)]
}

func elementType(
	expression ast.Expr,
) (element ast.Expr) {
	switch typed := expression.(type) {
	case *ast.StarExpr:
		return elementType(typed.X)
	case *ast.ParenExpr:
		return elementType(typed.X)
	case *ast.IndexExpr:
		return elementType(typed.X)
	case *ast.IndexListExpr:
		return elementType(typed.X)
	}

	return expression
}

func qualifiedName(
	expression ast.Expr,
) (name string) {
	selector, isSelector := elementType(expression).(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}

	return identifierName(selector.X) + methodSeparator + selector.Sel.Name
}

func structFields(
	structure *ast.StructType,
) (types map[string]ast.Expr) {
	types = map[string]ast.Expr{}
	for _, field := range structure.Fields.List {
		addField(
			types,
			field,
		)
	}

	return types
}

func addField(
	types map[string]ast.Expr,
	field *ast.Field,
) {
	for _, name := range fieldNames(field) {
		types[name] = field.Type
	}
}

func fieldNames(
	field *ast.Field,
) (names []string) {
	if len(field.Names) == 0 {
		return []string{embeddedName(field.Type)}
	}

	for _, name := range field.Names {
		names = append(
			names,
			name.Name,
		)
	}

	return names
}

func embeddedName(
	expression ast.Expr,
) (name string) {
	selector, isSelector := elementType(expression).(*ast.SelectorExpr)
	if isSelector {
		return selector.Sel.Name
	}

	return typeName(expression)
}
