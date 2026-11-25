package cqs

import (
	"go/ast"
	"maps"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func packageEffects(
	files []source.File,
) (effects map[string][]int) {
	callables := packageCallables(files)
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
			callableOf(declaration),
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
