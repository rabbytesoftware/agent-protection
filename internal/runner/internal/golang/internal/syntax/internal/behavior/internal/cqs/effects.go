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
	qualified map[string]ast.Expr,
) (fields map[string]ast.Expr) {
	fields = maps.Clone(qualified)
	for _, spec := range typeSpecs(files) {
		copyFields(
			fields,
			spec,
			spec.Name.Name,
		)
	}

	return fields
}

func qualifiedMethods(
	files []source.File,
) (methods map[string]ast.Expr) {
	methods = map[string]ast.Expr{}
	for _, file := range files {
		addQualifiedMethods(
			methods,
			file,
		)
	}

	return methods
}

func addQualifiedMethods(
	methods map[string]ast.Expr,
	file source.File,
) {
	for _, spec := range typeSpecs([]source.File{file}) {
		copyMethods(
			methods,
			spec,
			file.Syntax.Name.Name+methodSeparator+spec.Name.Name,
		)
	}
}

func typeSpecs(
	files []source.File,
) (specs []*ast.TypeSpec) {
	for _, file := range files {
		specs = append(
			specs,
			source.Nodes[*ast.TypeSpec](file.Syntax)...,
		)
	}

	return specs
}

func copyFields(
	fields map[string]ast.Expr,
	spec *ast.TypeSpec,
	owner string,
) {
	copyMethods(
		fields,
		spec,
		owner,
	)
	structure, isStruct := spec.Type.(*ast.StructType)
	if !isStruct {
		return
	}

	for name, declared := range structFields(structure) {
		fields[methodKey(
			owner,
			name,
		)] = declared
	}
}

func copyMethods(
	fields map[string]ast.Expr,
	spec *ast.TypeSpec,
	owner string,
) {
	contract, isInterface := spec.Type.(*ast.InterfaceType)
	if !isInterface {
		return
	}

	for _, method := range contract.Methods.List {
		addMethod(
			fields,
			method,
			owner,
		)
	}
}

func addMethod(
	fields map[string]ast.Expr,
	method *ast.Field,
	owner string,
) {
	signature, isSignature := method.Type.(*ast.FuncType)
	if !isSignature {
		return
	}

	for _, name := range method.Names {
		fields[methodKey(
			owner,
			name.Name,
		)] = signature
	}
}
