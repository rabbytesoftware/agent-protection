package deadstore

import (
	"go/ast"
	"go/token"
	"slices"
)

func visitCompound(
	tracker *flow,
	statement ast.Stmt,
) {
	switch typed := statement.(type) {
	case *ast.IfStmt:
		visitIf(
			tracker,
			typed,
		)
	case *ast.ForStmt:
		visitFor(
			tracker,
			typed,
		)
	case *ast.RangeStmt:
		visitRange(
			tracker,
			typed,
		)
	case *ast.SwitchStmt:
		visitSwitch(
			tracker,
			typed,
		)
	case *ast.TypeSwitchStmt:
		visitTypeSwitch(
			tracker,
			typed,
		)
	case *ast.SelectStmt:
		visitSelect(
			tracker,
			typed,
		)
	default:
		readAll(
			tracker,
			statement,
		)
	}
}

func visitIf(
	tracker *flow,
	statement *ast.IfStmt,
) {
	tracker.pushScope()
	visitOptional(
		tracker,
		statement.Init,
	)
	readAll(
		tracker,
		statement.Cond,
	)
	entry := tracker.current.clone()
	visitBlock(
		tracker,
		statement.Body.List,
	)
	taken := tracker.current
	tracker.current = entry
	visitOptional(
		tracker,
		statement.Else,
	)
	tracker.current = taken.union(tracker.current)
	tracker.popScope()
}

func visitSwitch(
	tracker *flow,
	statement *ast.SwitchStmt,
) {
	tracker.pushScope()
	visitOptional(
		tracker,
		statement.Init,
	)
	readAll(
		tracker,
		statement.Tag,
	)
	visitClauses(
		tracker,
		statement.Body.List,
		nil,
	)
	tracker.popScope()
}

func visitTypeSwitch(
	tracker *flow,
	statement *ast.TypeSwitchStmt,
) {
	tracker.pushScope()
	visitOptional(
		tracker,
		statement.Init,
	)
	readAll(
		tracker,
		statement.Assign,
	)
	visitClauses(
		tracker,
		statement.Body.List,
		guardName(statement.Assign),
	)
	tracker.popScope()
}

func guardName(
	guard ast.Stmt,
) (name *ast.Ident) {
	assignment, isAssignment := guard.(*ast.AssignStmt)
	if !isAssignment {
		return nil
	}

	name, isIdentifier := assignment.Lhs[0].(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	return name
}

func visitClauses(
	tracker *flow,
	clauses []ast.Stmt,
	guard *ast.Ident,
) {
	entry := tracker.current.clone()
	exits := &frame{}
	tracker.frames = append(
		tracker.frames,
		exits,
	)
	for _, clause := range clauses {
		visitCase(
			tracker,
			entry,
			clause,
			guard,
		)
	}

	if !slices.ContainsFunc(
		clauses,
		isDefaultCase,
	) {
		exits.breaks = append(
			exits.breaks,
			entry,
		)
	}

	tracker.frames = tracker.frames[:len(tracker.frames)-1]
	tracker.current = joined(append(
		exits.breaks,
		exits.carried,
	))
}

func visitCase(
	tracker *flow,
	entry reaching,
	clause ast.Stmt,
	guard *ast.Ident,
) {
	exits := tracker.frames[len(tracker.frames)-1]
	tracker.current = entry.union(exits.carried)
	exits.carried = nil
	for _, expression := range caseExpressions(clause) {
		readAll(
			tracker,
			expression,
		)
	}

	tracker.pushScope()
	tracker.declare(guard)
	visitList(
		tracker,
		clauseBody(clause),
	)
	tracker.popScope()
	if fallsThrough(clause) {
		exits.carried = tracker.current
		return
	}

	exits.breaks = append(
		exits.breaks,
		tracker.current,
	)
}

func clauseBody(
	clause ast.Stmt,
) (statements []ast.Stmt) {
	typed, isCase := clause.(*ast.CaseClause)
	if !isCase {
		return nil
	}

	return typed.Body
}

func caseExpressions(
	clause ast.Stmt,
) (expressions []ast.Expr) {
	typed, isCase := clause.(*ast.CaseClause)
	if !isCase {
		return nil
	}

	return typed.List
}

func isDefaultCase(
	clause ast.Stmt,
) (isDefault bool) {
	typed, isCase := clause.(*ast.CaseClause)
	return isCase && typed.List == nil
}

func fallsThrough(
	clause ast.Stmt,
) (falls bool) {
	body := clauseBody(clause)
	if len(body) == 0 {
		return false
	}

	branch, isBranch := body[len(body)-1].(*ast.BranchStmt)
	return isBranch && branch.Tok == token.FALLTHROUGH
}

func visitSelect(
	tracker *flow,
	statement *ast.SelectStmt,
) {
	entry := tracker.current.clone()
	exits := &frame{breaks: []reaching{entry}}
	tracker.frames = append(
		tracker.frames,
		exits,
	)
	for _, clause := range statement.Body.List {
		visitComm(
			tracker,
			entry,
			clause,
		)
	}

	tracker.frames = tracker.frames[:len(tracker.frames)-1]
	tracker.current = joined(exits.breaks)
}

func visitComm(
	tracker *flow,
	entry reaching,
	clause ast.Stmt,
) {
	exits := tracker.frames[len(tracker.frames)-1]
	tracker.current = entry.clone()
	communication, isCommunication := clause.(*ast.CommClause)
	if !isCommunication {
		return
	}

	tracker.pushScope()
	visitOptional(
		tracker,
		communication.Comm,
	)
	visitList(
		tracker,
		communication.Body,
	)
	tracker.popScope()
	exits.breaks = append(
		exits.breaks,
		tracker.current,
	)
}
