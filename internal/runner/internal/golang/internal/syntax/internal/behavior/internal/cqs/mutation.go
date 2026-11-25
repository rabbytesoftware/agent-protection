package cqs

import (
	"go/ast"
	"maps"
	"slices"
	"strings"
)

const (
	appendName   = "append"
	atomicPrefix = "atomic."
)

func directPositions(
	described callable,
) (positions []int) {
	for _, name := range writtenRoots(described.declaration.Body) {
		positions = appendPosition(
			positions,
			described.pointers,
			name,
		)
	}

	for _, name := range slices.Concat(
		sharedRoots(described.declaration.Body),
		stdlibRoots(described),
		commandRoots(described),
		pointerFieldRoots(described),
	) {
		positions = appendPosition(
			positions,
			everySlot(described),
			name,
		)
	}

	return positions
}

func stdlibRoots(
	described callable,
) (names []string) {
	for _, call := range stdlibCalls(described) {
		names = append(
			names,
			rootName(receiverOf(call.Fun)),
		)
	}

	return names
}

func stdlibCalls(
	described callable,
) (calls []*ast.CallExpr) {
	for _, call := range described.calls {
		calls = appendStdlibCall(
			calls,
			described,
			call,
		)
	}

	return calls
}

func appendStdlibCall(
	calls []*ast.CallExpr,
	described callable,
	call *ast.CallExpr,
) (extended []*ast.CallExpr) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return calls
	}

	mutators := stdlibMutators()[qualifiedName(resolvedType(
		described,
		selector.X,
	))]
	if !slices.Contains(
		mutators,
		selector.Sel.Name,
	) {
		return calls
	}

	return append(
		calls,
		call,
	)
}

func appendPosition(
	positions []int,
	slots map[string]int,
	name string,
) (extended []int) {
	position, isSlot := slots[name]
	if !isSlot {
		return positions
	}

	return append(
		positions,
		position,
	)
}

func writtenRoots(
	body *ast.BlockStmt,
) (names []string) {
	for _, target := range writeTargets(body) {
		names = append(
			names,
			mutatedRoot(target),
		)
	}

	return names
}

func pointerFieldRoots(
	described callable,
) (names []string) {
	for _, target := range writeTargets(described.declaration.Body) {
		names = appendPointerField(
			names,
			described,
			target,
		)
	}

	return names
}

func appendPointerField(
	names []string,
	described callable,
	target ast.Expr,
) (extended []string) {
	if !throughPointerField(
		described,
		target,
	) {
		return names
	}

	return append(
		names,
		rootName(target),
	)
}

func throughPointerField(
	described callable,
	target ast.Expr,
) (through bool) {
	selector, isSelector := ast.Unparen(target).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	_, isPointer := resolvedType(
		described,
		selector.X,
	).(*ast.StarExpr)
	return isPointer || throughPointerField(
		described,
		selector.X,
	)
}

func writeTargets(
	body *ast.BlockStmt,
) (targets []ast.Expr) {
	for _, assignment := range liveNodes[*ast.AssignStmt](body) {
		targets = append(
			targets,
			assignment.Lhs...,
		)
	}

	for _, increment := range liveNodes[*ast.IncDecStmt](body) {
		targets = append(
			targets,
			increment.X,
		)
	}

	for _, call := range slices.DeleteFunc(
		liveNodes[*ast.CallExpr](body),
		isNotBuiltinWrite,
	) {
		targets = append(
			targets,
			call.Args[0],
		)
	}

	return targets
}

func everySlot(
	described callable,
) (slots map[string]int) {
	slots = maps.Clone(described.pointers)
	maps.Copy(
		slots,
		described.references,
	)
	return slots
}

func sharedRoots(
	body *ast.BlockStmt,
) (names []string) {
	for _, assignment := range liveNodes[*ast.AssignStmt](body) {
		names = append(
			names,
			sharedAssignmentRoots(assignment)...,
		)
	}

	for _, increment := range liveNodes[*ast.IncDecStmt](body) {
		names = appendShared(
			names,
			increment.X,
			nil,
		)
	}

	for _, call := range slices.DeleteFunc(
		liveNodes[*ast.CallExpr](body),
		isNotBuiltinWrite,
	) {
		names = append(
			names,
			rootName(call.Args[0]),
		)
	}

	for _, send := range liveNodes[*ast.SendStmt](body) {
		names = append(
			names,
			rootName(send.Chan),
		)
	}

	return names
}

func sharedAssignmentRoots(
	assignment *ast.AssignStmt,
) (names []string) {
	for position, target := range assignment.Lhs {
		names = appendShared(
			names,
			target,
			valueFor(
				assignment.Rhs,
				position,
			),
		)
	}

	return names
}

func appendShared(
	names []string,
	target ast.Expr,
	value ast.Expr,
) (extended []string) {
	if !isSharedWrite(
		target,
		value,
	) {
		return names
	}

	return append(
		names,
		rootName(target),
	)
}

func isSharedWrite(
	target ast.Expr,
	value ast.Expr,
) (shared bool) {
	_, isField := ast.Unparen(target).(*ast.SelectorExpr)
	return throughReference(target) || isField && isAppendCall(value)
}

func throughReference(
	target ast.Expr,
) (through bool) {
	switch typed := target.(type) {
	case *ast.SelectorExpr:
		return throughReference(typed.X)
	case *ast.ParenExpr:
		return throughReference(typed.X)
	case *ast.StarExpr, *ast.IndexExpr:
		return true
	}

	return false
}

func isAppendCall(
	value ast.Expr,
) (appends bool) {
	call, isCall := value.(*ast.CallExpr)
	return isCall && identifierName(call.Fun) == appendName
}

func isNotBuiltinWrite(
	call *ast.CallExpr,
) (excluded bool) {
	name := calleeName(call)
	return len(call.Args) == 0 || !slices.Contains(
		builtinWrites(),
		name,
	) && !isAtomicWrite(name)
}

func isAtomicWrite(
	name string,
) (writes bool) {
	operation, isAtomic := strings.CutPrefix(
		name,
		atomicPrefix,
	)
	for _, prefix := range atomicMutators() {
		writes = writes || isAtomic && strings.HasPrefix(
			operation,
			prefix,
		)
	}

	return writes
}

func mutatedRoot(
	target ast.Expr,
) (name string) {
	_, isIdentifier := ast.Unparen(target).(*ast.Ident)
	if isIdentifier {
		return ""
	}

	return rootName(target)
}

func rootName(
	expression ast.Expr,
) (name string) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return rootName(typed.X)
	case *ast.StarExpr:
		return rootName(typed.X)
	case *ast.IndexExpr:
		return rootName(typed.X)
	case *ast.ParenExpr:
		return rootName(typed.X)
	case *ast.UnaryExpr:
		return rootName(typed.X)
	}

	return ""
}

func builtinWrites() (names []string) {
	return []string{
		"delete",
		"clear",
		"copy",
		"slices.Sort",
		"slices.SortFunc",
		"slices.SortStableFunc",
		"slices.Reverse",
		"sort.Sort",
		"sort.Stable",
		"sort.Ints",
		"sort.Strings",
		"sort.Float64s",
		"sort.Slice",
		"sort.SliceStable",
		"maps.Copy",
		"maps.DeleteFunc",
	}
}

func atomicMutators() (names []string) {
	return []string{
		"Add",
		"Store",
		"Swap",
		"CompareAndSwap",
		"And",
		"Or",
	}
}

func stdlibMutators() (methods map[string][]string) {
	atomics := atomicMutators()
	buffers := bufferMutators()
	return map[string][]string{
		"atomic.Bool":     atomics,
		"atomic.Int32":    atomics,
		"atomic.Int64":    atomics,
		"atomic.Uint32":   atomics,
		"atomic.Uint64":   atomics,
		"atomic.Uintptr":  atomics,
		"atomic.Pointer":  atomics,
		"atomic.Value":    atomics,
		"sync.Map":        syncMapMutators(),
		"bytes.Buffer":    buffers,
		"strings.Builder": buffers,
		"list.List":       listMutators(),
	}
}

func bufferMutators() (names []string) {
	return []string{
		"Write",
		"WriteString",
		"WriteByte",
		"WriteRune",
		"WriteTo",
		"ReadFrom",
		"Reset",
		"Grow",
		"Truncate",
	}
}

func syncMapMutators() (names []string) {
	return []string{
		"Store",
		"LoadOrStore",
		"LoadAndDelete",
		"Delete",
		"Swap",
		"CompareAndSwap",
		"CompareAndDelete",
		"Clear",
	}
}

func listMutators() (names []string) {
	return []string{
		"PushBack",
		"PushFront",
		"PushBackList",
		"PushFrontList",
		"InsertBefore",
		"InsertAfter",
		"MoveToFront",
		"MoveToBack",
		"MoveBefore",
		"MoveAfter",
		"Remove",
		"Init",
	}
}
