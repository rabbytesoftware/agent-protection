package cqs

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func returnsCreatedValue(
	described callable,
	effects map[string][]int,
) (created bool) {
	values := slices.DeleteFunc(
		returnedValues(described.declaration),
		isZeroLiteral,
	)
	stored := storedNames(
		described,
		effects,
	)
	created = len(values) > 0
	for _, value := range values {
		created = created && isCreatedValue(
			described,
			value,
			stored,
		)
	}

	return created
}

func returnedValues(
	declaration *ast.FuncDecl,
) (values []ast.Expr) {
	slots := flatParameters(declaration.Type.Results)
	for _, statement := range ownReturns(declaration.Body) {
		values = append(
			values,
			nonErrorResults(
				statement,
				slots,
			)...,
		)
	}

	return values
}

func nonErrorResults(
	statement *ast.ReturnStmt,
	slots []*ast.Field,
) (values []ast.Expr) {
	if len(statement.Results) != len(slots) {
		return statement.Results
	}

	for position, result := range statement.Results {
		values = appendNonError(
			values,
			result,
			slots[position],
		)
	}

	return values
}

func appendNonError(
	values []ast.Expr,
	result ast.Expr,
	slot *ast.Field,
) (extended []ast.Expr) {
	if !isNonErrorResult(slot) {
		return values
	}

	return append(
		values,
		result,
	)
}

func isZeroLiteral(
	value ast.Expr,
) (zero bool) {
	literal, isLiteral := value.(*ast.BasicLit)
	if isLiteral {
		return slices.Contains(
			[]string{`""`, "``", "0"},
			literal.Value,
		)
	}

	return slices.Contains(
		[]string{"nil", "false"},
		identifierName(value),
	)
}

func isCreatedValue(
	described callable,
	value ast.Expr,
	stored []string,
) (created bool) {
	name := identifierName(rootOf(value))
	isCandidate := name != "" &&
		slices.Contains(
			stored,
			name,
		) &&
		!slices.Contains(
			parameterNames(described.declaration),
			name,
		)
	if !isCandidate {
		return false
	}

	origins := originsOf(
		described.declaration.Body,
		name,
	)
	created = len(origins) > 0
	for _, origin := range origins {
		created = created && isFresh(
			described,
			origin,
		)
	}

	return created
}

func parameterNames(
	declaration *ast.FuncDecl,
) (names []string) {
	parameters := slices.Concat(
		flatParameters(declaration.Recv),
		flatParameters(declaration.Type.Params),
	)
	for _, parameter := range parameters {
		names = append(
			names,
			parameter.Names[0].Name,
		)
	}

	return names
}

func isFresh(
	described callable,
	origin ast.Expr,
) (fresh bool) {
	return isConstruction(origin) && (isBuilderCall(origin) || !readsPointer(
		described,
		origin,
	))
}

func isBuilderCall(
	origin ast.Expr,
) (builds bool) {
	call, isCall := ast.Unparen(origin).(*ast.CallExpr)
	if !isCall {
		return false
	}

	name := strings.ToLower(lastName(call.Fun))
	for _, prefix := range builderPrefixes() {
		builds = builds || strings.HasPrefix(
			name,
			prefix,
		)
	}

	return builds
}

func builderPrefixes() (prefixes []string) {
	return []string{
		"new",
		"build",
		"make",
		"create",
		"prepare",
		"compose",
		"assemble",
		"construct",
	}
}

func rootOf(
	value ast.Expr,
) (root ast.Expr) {
	selector, isSelector := ast.Unparen(value).(*ast.SelectorExpr)
	if !isSelector {
		return value
	}

	return rootOf(selector.X)
}

func isConstruction(
	origin ast.Expr,
) (constructs bool) {
	switch typed := ast.Unparen(origin).(type) {
	case *ast.CallExpr, *ast.CompositeLit:
		return true
	case *ast.UnaryExpr:
		return typed.Op == token.AND
	}

	return false
}

func readsPointer(
	described callable,
	expression ast.Expr,
) (reads bool) {
	for _, identifier := range source.Nodes[*ast.Ident](expression) {
		_, isPointer := described.pointers[identifier.Name]
		reads = reads || isPointer
	}

	return reads
}

func storedNames(
	described callable,
	effects map[string][]int,
) (names []string) {
	written := []ast.Node{}
	for _, call := range slices.Concat(
		mutatingCalls(
			described,
			effects,
		),
		stdlibCalls(described),
		commandCalls(described),
	) {
		written = append(
			written,
			call,
		)
	}

	for _, assignment := range liveNodes[*ast.AssignStmt](described.declaration.Body) {
		written = appendPointerWrite(
			written,
			described,
			assignment,
		)
	}

	for _, node := range written {
		names = append(
			names,
			identifierNames(node)...,
		)
	}

	return names
}

func appendPointerWrite(
	written []ast.Node,
	described callable,
	assignment *ast.AssignStmt,
) (extended []ast.Node) {
	if !writesPointer(
		described,
		assignment,
	) {
		return written
	}

	return append(
		written,
		assignment,
	)
}

func writesPointer(
	described callable,
	assignment *ast.AssignStmt,
) (writes bool) {
	for _, target := range assignment.Lhs {
		_, isPointer := described.pointers[mutatedRoot(target)]
		writes = writes || isPointer || throughPointerField(
			described,
			target,
		)
	}

	shared := everySlot(described)
	for _, name := range sharedAssignmentRoots(assignment) {
		_, isShared := shared[name]
		writes = writes || isShared
	}

	return writes
}

func identifierNames(
	node ast.Node,
) (names []string) {
	for _, identifier := range source.Nodes[*ast.Ident](node) {
		names = append(
			names,
			identifier.Name,
		)
	}

	return names
}

func originsOf(
	body *ast.BlockStmt,
	name string,
) (origins []ast.Expr) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](body) {
		origins = append(
			origins,
			assignedValues(
				assignment.Lhs,
				assignment.Rhs,
				name,
			)...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](body) {
		origins = append(
			origins,
			specValues(
				spec,
				name,
			)...,
		)
	}

	return origins
}

func specValues(
	spec *ast.ValueSpec,
	name string,
) (values []ast.Expr) {
	targets := []ast.Expr{}
	for _, identifier := range spec.Names {
		targets = append(
			targets,
			identifier,
		)
	}

	return assignedValues(
		targets,
		spec.Values,
		name,
	)
}

func assignedValues(
	targets []ast.Expr,
	values []ast.Expr,
	name string,
) (assigned []ast.Expr) {
	for position, target := range targets {
		assigned = appendAssigned(
			assigned,
			target,
			name,
			valueFor(
				values,
				position,
			),
		)
	}

	return assigned
}

func appendAssigned(
	assigned []ast.Expr,
	target ast.Expr,
	name string,
	value ast.Expr,
) (extended []ast.Expr) {
	if identifierName(target) != name {
		return assigned
	}

	return append(
		assigned,
		value,
	)
}

func valueFor(
	values []ast.Expr,
	position int,
) (assigned ast.Expr) {
	if len(values) == 1 {
		return values[0]
	}

	if position < len(values) {
		return values[position]
	}

	return &ast.BadExpr{}
}
