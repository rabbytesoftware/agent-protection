package deadstore

import (
	"go/ast"
	"go/token"
)

func visitFor(
	tracker *flow,
	statement *ast.ForStmt,
) {
	tracker.pushScope()
	visitOptional(
		tracker,
		statement.Init,
	)
	iterate(
		tracker,
		statement,
		tracker.current.clone(),
	)
	tracker.popScope()
}

func visitRange(
	tracker *flow,
	statement *ast.RangeStmt,
) {
	readAll(
		tracker,
		statement.X,
	)
	tracker.pushScope()
	iterate(
		tracker,
		statement,
		tracker.current.clone(),
	)
	tracker.popScope()
}

func iterate(
	tracker *flow,
	loop ast.Stmt,
	head reaching,
) {
	exits := &frame{loop: loop}
	tracker.frames = append(
		tracker.frames,
		exits,
	)
	tracker.loops = append(
		tracker.loops,
		loop,
	)
	tracker.current = head.clone()
	enter(
		tracker,
		loop,
	)
	visitBlock(
		tracker,
		loopBody(loop).List,
	)
	tracker.current = joined(append(
		exits.continues,
		tracker.current,
	))
	tracker.markCycled(loop)
	visitOptional(
		tracker,
		loopPost(loop),
	)
	tracker.loops = tracker.loops[:len(tracker.loops)-1]
	tracker.frames = tracker.frames[:len(tracker.frames)-1]
	settle(
		tracker,
		loop,
		head,
		exits,
	)
}

func settle(
	tracker *flow,
	loop ast.Stmt,
	head reaching,
	exits *frame,
) {
	next := head.union(tracker.current)
	if !next.equal(head) {
		iterate(
			tracker,
			loop,
			next,
		)
		return
	}

	if mayEnd(loop) {
		exits.breaks = append(
			exits.breaks,
			head,
		)
	}

	tracker.current = joined(exits.breaks)
}

func enter(
	tracker *flow,
	loop ast.Stmt,
) {
	switch typed := loop.(type) {
	case *ast.ForStmt:
		readAll(
			tracker,
			typed.Cond,
		)
	case *ast.RangeStmt:
		bindRange(
			tracker,
			typed,
		)
	}
}

func bindRange(
	tracker *flow,
	statement *ast.RangeStmt,
) {
	for _, target := range []ast.Expr{
		statement.Key,
		statement.Value,
	} {
		bindRangeTarget(
			tracker,
			target,
			statement.Tok,
		)
	}
}

func bindRangeTarget(
	tracker *flow,
	target ast.Expr,
	operator token.Token,
) {
	identifier, isIdentifier := target.(*ast.Ident)
	if !isIdentifier {
		readAll(
			tracker,
			target,
		)
		return
	}

	if operator == token.DEFINE {
		tracker.declare(identifier)
	}

	tracker.assign(
		identifier,
		nil,
	)
}

func loopBody(
	loop ast.Stmt,
) (body *ast.BlockStmt) {
	switch typed := loop.(type) {
	case *ast.ForStmt:
		return typed.Body
	case *ast.RangeStmt:
		return typed.Body
	}

	return &ast.BlockStmt{}
}

func loopPost(
	loop ast.Stmt,
) (post ast.Stmt) {
	statement, isFor := loop.(*ast.ForStmt)
	if !isFor {
		return nil
	}

	return statement.Post
}

func mayEnd(
	loop ast.Stmt,
) (ends bool) {
	statement, isFor := loop.(*ast.ForStmt)
	return !isFor || statement.Cond != nil
}
