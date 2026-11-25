package publicapi

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	noLogic     = ""
	nilName     = "nil"
	logicFormat = "container.go func %s %s: entry functions only delegate; move the logic into an internal package"
)

func delegationFindings(
	file source.File,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	guards := guardConditions(function.Body)
	for _, node := range outermost(logicNodes(
		function.Body,
		guards,
	)) {
		findings = append(
			findings,
			report.Finding{
				Rule:     apiRule,
				Location: file.Location(node),
				Snippet: fmt.Sprintf(
					logicFormat,
					function.Name.Name,
					logicVerb(
						node,
						guards,
					),
				),
			},
		)
	}

	return findings
}

func guardConditions(
	body *ast.BlockStmt,
) (conditions []ast.Expr) {
	for _, statement := range slices.DeleteFunc(
		source.Nodes[*ast.IfStmt](body),
		isNotErrorGuard,
	) {
		conditions = append(
			conditions,
			ast.Unparen(statement.Cond),
		)
	}

	return conditions
}

func logicNodes(
	body *ast.BlockStmt,
	guards []ast.Expr,
) (nodes []ast.Node) {
	for node := range ast.Preorder(body) {
		nodes = append(
			nodes,
			logicNode(
				node,
				guards,
			)...,
		)
	}

	return nodes
}

func logicNode(
	node ast.Node,
	guards []ast.Expr,
) (logic []ast.Node) {
	if logicVerb(
		node,
		guards,
	) == noLogic {
		return nil
	}

	return []ast.Node{node}
}

func outermost(
	nodes []ast.Node,
) (kept []ast.Node) {
	for _, node := range nodes {
		kept = append(
			kept,
			uncovered(
				kept,
				node,
			)...,
		)
	}

	return kept
}

func uncovered(
	kept []ast.Node,
	node ast.Node,
) (fresh []ast.Node) {
	isCovered := len(kept) > 0 && node.Pos() < kept[len(kept)-1].End()
	if isCovered {
		return nil
	}

	return []ast.Node{node}
}

func logicVerb(
	node ast.Node,
	guards []ast.Expr,
) (verb string) {
	switch typed := node.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		return "loops"
	case *ast.SwitchStmt, *ast.TypeSwitchStmt:
		return "switches"
	case *ast.SelectStmt:
		return "selects"
	case *ast.GoStmt:
		return "starts a goroutine"
	case *ast.DeferStmt:
		return "defers"
	case *ast.FuncLit:
		return "declares a func literal"
	case *ast.IfStmt:
		return branchVerb(typed)
	}

	return computationVerb(
		node,
		guards,
	)
}

func branchVerb(
	statement *ast.IfStmt,
) (verb string) {
	if isNotErrorGuard(statement) {
		return "branches"
	}

	return noLogic
}

func computationVerb(
	node ast.Node,
	guards []ast.Expr,
) (verb string) {
	switch typed := node.(type) {
	case *ast.BinaryExpr:
		return guardVerb(
			typed,
			guards,
		)
	case *ast.UnaryExpr:
		return unaryVerb(typed)
	case *ast.AssignStmt:
		return assignmentVerb(typed)
	case *ast.IncDecStmt:
		return "computes"
	}

	return noLogic
}

func guardVerb(
	expression *ast.BinaryExpr,
	guards []ast.Expr,
) (verb string) {
	if slices.Contains(
		guards,
		ast.Expr(expression),
	) {
		return noLogic
	}

	return "computes"
}

func unaryVerb(
	expression *ast.UnaryExpr,
) (verb string) {
	isComputation := expression.Op == token.NOT || expression.Op == token.SUB || expression.Op == token.XOR
	if isComputation {
		return "computes"
	}

	return noLogic
}

func assignmentVerb(
	statement *ast.AssignStmt,
) (verb string) {
	isPlain := statement.Tok == token.ASSIGN || statement.Tok == token.DEFINE
	if isPlain {
		return noLogic
	}

	return "computes"
}

func isNotErrorGuard(
	statement *ast.IfStmt,
) (excluded bool) {
	condition, isBinary := ast.Unparen(statement.Cond).(*ast.BinaryExpr)
	isNilCheck := isBinary && condition.Op == token.NEQ && (isNil(condition.X) || isNil(condition.Y))
	isSingleReturn := len(statement.Body.List) == 1 && isReturn(statement.Body.List[0])
	return !isNilCheck || statement.Else != nil || !isSingleReturn
}

func isNil(
	expression ast.Expr,
) (matches bool) {
	identifier, isIdentifier := ast.Unparen(expression).(*ast.Ident)
	return isIdentifier && identifier.Name == nilName
}

func isReturn(
	statement ast.Stmt,
) (matches bool) {
	_, matches = statement.(*ast.ReturnStmt)
	return matches
}
