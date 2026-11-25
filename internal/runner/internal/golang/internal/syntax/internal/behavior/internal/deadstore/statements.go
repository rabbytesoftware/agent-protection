package deadstore

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	blankName  = "_"
	callSuffix = "(...)"
)

func visit(
	tracker *flow,
	statement ast.Stmt,
) {
	switch typed := statement.(type) {
	case *ast.AssignStmt:
		visitAssign(
			tracker,
			typed,
		)
	case *ast.DeclStmt:
		visitDecl(
			tracker,
			typed,
		)
	case *ast.ReturnStmt:
		visitReturn(
			tracker,
			typed,
		)
	case *ast.BranchStmt:
		visitBranch(
			tracker,
			typed,
		)
	case *ast.BlockStmt:
		visitBlock(
			tracker,
			typed.List,
		)
	default:
		visitCompound(
			tracker,
			statement,
		)
	}
}

func visitOptional(
	tracker *flow,
	statement ast.Stmt,
) {
	if statement == nil {
		return
	}

	visit(
		tracker,
		statement,
	)
}

func visitBlock(
	tracker *flow,
	statements []ast.Stmt,
) {
	tracker.pushScope()
	visitList(
		tracker,
		statements,
	)
	tracker.popScope()
}

func visitList(
	tracker *flow,
	statements []ast.Stmt,
) {
	for _, statement := range statements {
		visit(
			tracker,
			statement,
		)
	}
}

func visitAssign(
	tracker *flow,
	assignment *ast.AssignStmt,
) {
	for _, value := range assignment.Rhs {
		readAll(
			tracker,
			value,
		)
	}

	for _, target := range assignment.Lhs {
		readTarget(
			tracker,
			target,
			assignment.Tok,
		)
	}

	for position, target := range assignment.Lhs {
		bind(
			tracker,
			target,
			assignment,
			position,
		)
	}
}

func readTarget(
	tracker *flow,
	target ast.Expr,
	operator token.Token,
) {
	_, isIdentifier := target.(*ast.Ident)
	isPlainWrite := isIdentifier && (operator == token.ASSIGN || operator == token.DEFINE)
	if isPlainWrite {
		return
	}

	readAll(
		tracker,
		target,
	)
}

func bind(
	tracker *flow,
	target ast.Expr,
	assignment *ast.AssignStmt,
	position int,
) {
	identifier, isIdentifier := target.(*ast.Ident)
	if !isIdentifier ||
		identifier.Name == blankName {
		return
	}

	isDeclaration := assignment.Tok == token.DEFINE && !tracker.declaredHere(identifier.Name)
	if isDeclaration {
		tracker.declare(identifier)
	}

	tracker.assign(
		identifier,
		valueFor(
			assignment.Rhs,
			position,
		),
	)
}

func visitDecl(
	tracker *flow,
	declaration *ast.DeclStmt,
) {
	for _, spec := range source.Nodes[*ast.ValueSpec](declaration) {
		visitSpec(
			tracker,
			spec,
		)
	}
}

func visitSpec(
	tracker *flow,
	spec *ast.ValueSpec,
) {
	for _, value := range spec.Values {
		readAll(
			tracker,
			value,
		)
	}

	for position, name := range spec.Names {
		tracker.declare(name)
		bindSpec(
			tracker,
			spec,
			name,
			position,
		)
	}
}

func bindSpec(
	tracker *flow,
	spec *ast.ValueSpec,
	name *ast.Ident,
	position int,
) {
	if len(spec.Values) == 0 {
		return
	}

	tracker.assign(
		name,
		valueFor(
			spec.Values,
			position,
		),
	)
}

func visitReturn(
	tracker *flow,
	statement *ast.ReturnStmt,
) {
	for _, result := range statement.Results {
		readAll(
			tracker,
			result,
		)
	}

	for _, variable := range nakedResults(
		tracker,
		statement,
	) {
		tracker.readVariable(
			variable,
			statement.Pos(),
		)
	}

	tracker.exit()
}

func nakedResults(
	tracker *flow,
	statement *ast.ReturnStmt,
) (variables []token.Pos) {
	if len(statement.Results) > 0 {
		return nil
	}

	return tracker.results
}

func visitBranch(
	tracker *flow,
	statement *ast.BranchStmt,
) {
	if statement.Tok == token.BREAK {
		tracker.breakOut()
		return
	}

	if statement.Tok == token.CONTINUE {
		tracker.continueLoop()
	}
}

func readAll(
	tracker *flow,
	node ast.Node,
) {
	if node == nil {
		return
	}

	selected := []*ast.Ident{}
	for _, selector := range source.Nodes[*ast.SelectorExpr](node) {
		selected = append(
			selected,
			selector.Sel,
		)
	}

	for _, identifier := range source.Nodes[*ast.Ident](node) {
		readUnselected(
			tracker,
			identifier,
			selected,
		)
	}
}

func readUnselected(
	tracker *flow,
	identifier *ast.Ident,
	selected []*ast.Ident,
) {
	if slices.Contains(
		selected,
		identifier,
	) {
		return
	}

	tracker.read(identifier)
}

func valueFor(
	values []ast.Expr,
	position int,
) (value ast.Expr) {
	if len(values) == 1 {
		return values[0]
	}

	if position < len(values) {
		return values[position]
	}

	return nil
}

func callName(
	value ast.Expr,
) (name string) {
	call, isCall := ast.Unparen(value).(*ast.CallExpr)
	if !isCall {
		return ""
	}

	return types.ExprString(call.Fun) + callSuffix
}
