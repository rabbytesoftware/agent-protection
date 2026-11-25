package interfacesize

import (
	"go/ast"
	"slices"
)

func methodNames(
	declared declaration,
	index map[string]declaration,
	visiting []string,
) (names []string) {
	path := slices.Concat(
		visiting,
		[]string{declarationKey(
			declared.directory,
			declared.spec.Name.Name,
		)},
	)
	for _, field := range declared.body.Methods.List {
		names = append(
			names,
			fieldMethods(
				declared,
				field,
				index,
				path,
			)...,
		)
	}

	slices.Sort(names)
	return slices.Compact(names)
}

func fieldMethods(
	declared declaration,
	field *ast.Field,
	index map[string]declaration,
	visiting []string,
) (names []string) {
	for _, name := range field.Names {
		names = append(
			names,
			name.Name,
		)
	}

	key, isResolved := embeddedKey(
		declared,
		field,
		index,
	)
	isExpandable := isResolved && !slices.Contains(
		visiting,
		key,
	)
	if !isExpandable {
		return names
	}

	return append(
		names,
		methodNames(
			index[key],
			index,
			visiting,
		)...,
	)
}

func resolvedEmbeds(
	declared declaration,
	index map[string]declaration,
) (embeds []*ast.Field) {
	for _, field := range declared.body.Methods.List {
		embeds = appendResolved(
			embeds,
			declared,
			field,
			index,
		)
	}

	return embeds
}

func appendResolved(
	embeds []*ast.Field,
	declared declaration,
	field *ast.Field,
	index map[string]declaration,
) (extended []*ast.Field) {
	_, isResolved := embeddedKey(
		declared,
		field,
		index,
	)
	if !isResolved {
		return embeds
	}

	return append(
		embeds,
		field,
	)
}

func embeddedKey(
	declared declaration,
	field *ast.Field,
	index map[string]declaration,
) (key string, isResolved bool) {
	if len(field.Names) > 0 {
		return "",
			false
	}

	key = memberKey(
		declared,
		unindexed(field.Type),
	)
	_, isResolved = index[key]
	return key,
		isResolved
}

func memberKey(
	declared declaration,
	member ast.Expr,
) (key string) {
	switch typed := member.(type) {
	case *ast.Ident:
		return declarationKey(
			declared.directory,
			typed.Name,
		)
	case *ast.SelectorExpr:
		return qualifiedKey(
			declared,
			typed,
		)
	}

	return ""
}

func qualifiedKey(
	declared declaration,
	selector *ast.SelectorExpr,
) (key string) {
	qualifier, isIdentifier := selector.X.(*ast.Ident)
	if !isIdentifier {
		return ""
	}

	directory, isImported := declared.aliases[qualifier.Name]
	if !isImported {
		return ""
	}

	return declarationKey(
		directory,
		selector.Sel.Name,
	)
}

func unindexed(
	expression ast.Expr,
) (inner ast.Expr) {
	switch typed := expression.(type) {
	case *ast.IndexExpr:
		return typed.X
	case *ast.IndexListExpr:
		return typed.X
	}

	return expression
}
