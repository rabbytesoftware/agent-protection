package assertion

import (
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func provenDeclarations(
	known catalog,
) (proven []int) {
	edges := make(
		[][]int,
		len(known.declarations),
	)
	for index, candidate := range known.declarations {
		edges[index] = callees(
			known,
			candidate,
		)
		proven = append(
			proven,
			directProof(
				known,
				index,
				candidate,
			)...,
		)
	}

	for grown := true; grown; {
		proven, grown = expanded(
			proven,
			edges,
		)
	}

	return proven
}

func directProof(
	known catalog,
	index int,
	candidate declaration,
) (proven []int) {
	if !assertsDirectly(
		known,
		candidate,
	) {
		return nil
	}

	return []int{index}
}

func expanded(
	proven []int,
	edges [][]int,
) (grownProof []int, grown bool) {
	grownProof = slices.Clone(proven)
	for index, targets := range edges {
		grownProof = append(
			grownProof,
			newlyProven(
				index,
				targets,
				grownProof,
			)...,
		)
	}

	return grownProof,
		len(grownProof) > len(proven)
}

func newlyProven(
	index int,
	targets []int,
	proven []int,
) (added []int) {
	isNewlyProven := !slices.Contains(
		proven,
		index,
	) && reachesProof(
		targets,
		proven,
	)
	if !isNewlyProven {
		return nil
	}

	return []int{index}
}

func reachesProof(
	targets []int,
	proven []int,
) (reaches bool) {
	for _, target := range targets {
		reaches = reaches || slices.Contains(
			proven,
			target,
		)
	}

	return reaches
}

func assertsDirectly(
	known catalog,
	owner declaration,
) (asserts bool) {
	names := testingNames(
		owner,
		known.helpers,
	)
	fields := known.fields[filepath.Dir(owner.file.Rel)]
	for _, call := range source.Nodes[*ast.CallExpr](owner.body) {
		asserts = asserts || isFailingCall(
			call,
			names,
			fields,
		) || isLibraryAssertion(
			owner.imports,
			call,
			names,
		)
	}

	return asserts
}

func isFailingCall(
	call *ast.CallExpr,
	names []string,
	fields []string,
) (failing bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector ||
		!slices.Contains(
			failingMethods(),
			selector.Sel.Name,
		) {
		return false
	}

	return isTestingReceiver(
		selector.X,
		names,
		fields,
	)
}

func isTestingReceiver(
	receiver ast.Expr,
	names []string,
	fields []string,
) (testing bool) {
	switch typed := ast.Unparen(receiver).(type) {
	case *ast.Ident:
		return isTestingValue(
			typed,
			names,
		)
	case *ast.SelectorExpr:
		return slices.Contains(
			fields,
			typed.Sel.Name,
		)
	case *ast.CallExpr:
		return isSuiteTesting(typed)
	}

	return false
}

func isSuiteTesting(
	call *ast.CallExpr,
) (suite bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	return isSelector && len(call.Args) == 0 && selector.Sel.Name == suiteTesting
}

func callees(
	known catalog,
	owner declaration,
) (targets []int) {
	names := testingNames(
		owner,
		known.helpers,
	)
	for _, call := range source.Nodes[*ast.CallExpr](owner.body) {
		targets = append(
			targets,
			callTargets(
				known,
				owner,
				call,
				names,
			)...,
		)
	}

	return append(
		targets,
		referencedTargets(
			known,
			owner,
		)...,
	)
}

func referencedTargets(
	known catalog,
	owner declaration,
) (targets []int) {
	called := calledNames(owner.body)
	for _, reference := range references(owner.body) {
		targets = append(
			targets,
			referenceTargets(
				known,
				owner,
				called,
				reference,
			)...,
		)
	}

	return targets
}

func referenceTargets(
	known catalog,
	owner declaration,
	called []*ast.Ident,
	reference ast.Expr,
) (targets []int) {
	if slices.Contains(
		called,
		referenceName(reference),
	) {
		return nil
	}

	return untestedTargets(
		known,
		known.resolve(
			owner,
			reference,
		),
	)
}

func untestedTargets(
	known catalog,
	candidates []int,
) (targets []int) {
	for _, candidate := range candidates {
		targets = append(
			targets,
			untestedTarget(
				known,
				candidate,
			)...,
		)
	}

	return targets
}

func untestedTarget(
	known catalog,
	candidate int,
) (targets []int) {
	if takesTesting(
		known,
		known.declarations[candidate],
	) {
		return nil
	}

	return []int{candidate}
}

func references(
	body *ast.BlockStmt,
) (found []ast.Expr) {
	for _, selector := range source.Nodes[*ast.SelectorExpr](body) {
		found = append(
			found,
			selector,
		)
	}

	for _, identifier := range source.Nodes[*ast.Ident](body) {
		found = append(
			found,
			functionReference(identifier)...,
		)
	}

	return found
}

func functionReference(
	identifier *ast.Ident,
) (found []ast.Expr) {
	isFunctionReference := identifier.Obj == nil || identifier.Obj.Kind == ast.Fun
	if !isFunctionReference {
		return nil
	}

	return []ast.Expr{identifier}
}

func referenceName(
	reference ast.Expr,
) (name *ast.Ident) {
	selector, isSelector := reference.(*ast.SelectorExpr)
	if isSelector {
		return selector.Sel
	}

	identifier, isIdent := reference.(*ast.Ident)
	if !isIdent {
		return nil
	}

	return identifier
}

func callTargets(
	known catalog,
	owner declaration,
	call *ast.CallExpr,
	names []string,
) (targets []int) {
	targets = calledTargets(
		known,
		owner,
		call,
		names,
	)
	isDriven := passesTesting(
		call.Args,
		names,
	) || isDrivenByTesting(
		call,
		names,
	)
	if !isDriven {
		return targets
	}

	return append(
		targets,
		functionValues(
			known,
			owner,
			call.Args,
		)...,
	)
}

func calledTargets(
	known catalog,
	owner declaration,
	call *ast.CallExpr,
	names []string,
) (targets []int) {
	resolved := known.resolve(
		owner,
		call.Fun,
	)
	if !passesTesting(
		call.Args,
		names,
	) {
		return untestedTargets(
			known,
			resolved,
		)
	}

	if len(resolved) > 0 ||
		isImported(
			owner,
			call.Fun,
		) {
		return resolved
	}

	return known.dynamicTargets(owner)
}

func isDrivenByTesting(
	call *ast.CallExpr,
	names []string,
) (driven bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	return isSelector && isTestingValue(
		selector.X,
		names,
	)
}

func functionValues(
	known catalog,
	owner declaration,
	arguments []ast.Expr,
) (targets []int) {
	for _, argument := range arguments {
		targets = append(
			targets,
			known.resolve(
				owner,
				argument,
			)...,
		)
	}

	return targets
}

func takesTesting(
	known catalog,
	target declaration,
) (takes bool) {
	return len(fieldNames(
		target.function.Params,
		testingAlias(target.imports),
		known.helpers,
	)) > 0
}

func isImported(
	owner declaration,
	function ast.Expr,
) (imported bool) {
	selector, isSelector := unwrapped(function).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	_, imported = owner.imports[identifierName(qualifier)]
	return isIdent && imported
}
