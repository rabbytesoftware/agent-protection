package cqs

import (
	"go/ast"
	"maps"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func packageEffects(
	files []source.File,
	fields map[string]ast.Expr,
) (effects map[string][]int) {
	callables := packageCallables(
		files,
		fields,
	)
	direct := map[string][]int{}
	for _, described := range callables {
		direct[described.key] = merged(
			direct[described.key],
			directPositions(described),
		)
	}

	return settled(
		callables,
		direct,
	)
}

func packageCallables(
	files []source.File,
	fields map[string]ast.Expr,
) (callables []callable) {
	declarations := []*ast.FuncDecl{}
	for _, file := range files {
		declarations = append(
			declarations,
			source.Nodes[*ast.FuncDecl](file.Syntax)...,
		)
	}

	for _, declaration := range slices.DeleteFunc(
		declarations,
		hasNoBody,
	) {
		callables = append(
			callables,
			callableOf(
				declaration,
				fields,
			),
		)
	}

	return callables
}

func settled(
	callables []callable,
	effects map[string][]int,
) (fixpoint map[string][]int) {
	next := spread(
		callables,
		effects,
	)
	if maps.EqualFunc(
		next,
		effects,
		slices.Equal[[]int],
	) {
		return effects
	}

	return settled(
		callables,
		next,
	)
}

func spread(
	callables []callable,
	effects map[string][]int,
) (next map[string][]int) {
	next = maps.Clone(effects)
	for _, described := range callables {
		next[described.key] = merged(
			next[described.key],
			callPositions(
				described,
				effects,
			),
		)
	}

	return next
}

func merged(
	positions []int,
	more []int,
) (union []int) {
	return slices.Compact(slices.Sorted(slices.Values(slices.Concat(
		positions,
		more,
	))))
}

func hasNoBody(
	declaration *ast.FuncDecl,
) (bodiless bool) {
	return declaration.Body == nil
}

func packageFields(
	files []source.File,
) (fields map[string]ast.Expr) {
	fields = map[string]ast.Expr{}
	specs := []*ast.TypeSpec{}
	for _, file := range files {
		specs = append(
			specs,
			source.Nodes[*ast.TypeSpec](file.Syntax)...,
		)
	}

	for _, spec := range specs {
		copyFields(
			fields,
			spec,
		)
	}

	return fields
}

func copyFields(
	fields map[string]ast.Expr,
	spec *ast.TypeSpec,
) {
	structure, isStruct := spec.Type.(*ast.StructType)
	if !isStruct {
		return
	}

	for name, declared := range structFields(structure) {
		fields[methodKey(
			spec.Name.Name,
			name,
		)] = declared
	}
}
