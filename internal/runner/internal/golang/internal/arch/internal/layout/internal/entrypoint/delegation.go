package entrypoint

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	noLogic     = ""
	logicFormat = "%s func %s %s: entry functions only delegate; %s"
	logicRemedy = "move the logic into an internal package"
	callRemedy  = "call only this module's internal packages, its own declarations, errors.New, errors.Join, fmt.Errorf, slices.Concat or other modules' constructors, and take configuration as parameters"
	wireRemedy  = "call another module's constructor only to assign its result and pass it on or return it; flags, environment and configuration files are read by the caller and arrive as entry parameters"
	fieldRemedy = "a container passes dependencies to constructors and returns what they build; construct the object in an internal package and return it as this module's interface instead of filling its fields here"
	fieldVerb   = "sets field "
)

type delegation struct {
	unit        unit
	guards      []ast.Expr
	locals      []string
	typeIndexes []ast.Node
	wiring      []ast.Expr
}

func delegationFindings(
	found unit,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	context := delegation{
		unit: found,
		guards: guardConditions(
			function.Body,
			found.role,
		),
		locals: localNames(function),
		typeIndexes: typeIndexes(
			found,
			function.Body,
		),
		wiring: wiringCalls(function.Body),
	}
	for _, node := range outermost(
		context,
		logicNodes(
			context,
			function.Body,
		),
	) {
		findings = append(
			findings,
			logicFinding(
				context,
				function,
				node,
			),
		)
	}

	return findings
}

func logicFinding(
	context delegation,
	function *ast.FuncDecl,
	node ast.Node,
) (finding report.Finding) {
	verb := logicVerb(
		context,
		node,
	)
	return report.Finding{
		Rule:     apiRule,
		Location: context.unit.file.Location(node),
		Snippet: fmt.Sprintf(
			logicFormat,
			filepath.Base(context.unit.file.Rel),
			function.Name.Name,
			verb,
			remedyOf(
				node,
				verb,
			),
		),
	}
}

func remedyOf(
	node ast.Node,
	verb string,
) (remedy string) {
	isForeign := strings.HasPrefix(
		verb,
		configurationVerb,
	) || strings.HasPrefix(
		verb,
		libraryVerb,
	) || strings.HasSuffix(
		verb,
		unwiredVerb,
	)
	_, isCall := node.(*ast.CallExpr)
	switch {
	case isForeign:
		return wireRemedy
	case strings.HasPrefix(
		verb,
		fieldVerb,
	):
		return fieldRemedy
	case isCall:
		return callRemedy
	}

	return logicRemedy
}

func logicNodes(
	context delegation,
	body *ast.BlockStmt,
) (nodes []ast.Node) {
	for node := range ast.Preorder(body) {
		nodes = append(
			nodes,
			logicNode(
				context,
				node,
			)...,
		)
	}

	return nodes
}

func logicNode(
	context delegation,
	node ast.Node,
) (logic []ast.Node) {
	if logicVerb(
		context,
		node,
	) == noLogic {
		return nil
	}

	return []ast.Node{node}
}

func outermost(
	context delegation,
	nodes []ast.Node,
) (kept []ast.Node) {
	for _, node := range nodes {
		kept = append(
			kept,
			uncovered(
				context,
				kept,
				node,
			)...,
		)
	}

	return kept
}

func uncovered(
	context delegation,
	kept []ast.Node,
	node ast.Node,
) (fresh []ast.Node) {
	isCovered := false
	for _, outer := range kept {
		isCovered = isCovered || covers(
			context,
			outer,
			node,
		)
	}

	if isCovered {
		return nil
	}

	return []ast.Node{node}
}

func covers(
	context delegation,
	outer ast.Node,
	node ast.Node,
) (covered bool) {
	_, isStatement := outer.(ast.Stmt)
	_, isLiteral := outer.(*ast.FuncLit)
	isInside := outer.Pos() <= node.Pos() && node.End() <= outer.End()
	isSameLogic := isStatement || isLiteral || logicVerb(
		context,
		outer,
	) == logicVerb(
		context,
		node,
	)
	return isInside && isSameLogic
}

func logicVerb(
	context delegation,
	node ast.Node,
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
		return branchVerb(
			context,
			typed,
		)
	}

	return expressionVerb(
		context,
		node,
	)
}

func expressionVerb(
	context delegation,
	node ast.Node,
) (verb string) {
	switch typed := node.(type) {
	case *ast.CallExpr:
		return callVerb(
			context,
			typed,
		)
	case *ast.IndexExpr:
		return indexVerb(
			context,
			typed,
		)
	case *ast.CompositeLit:
		return compositeVerb(
			context,
			typed,
		)
	}

	return computationVerb(
		context,
		node,
	)
}

func branchVerb(
	context delegation,
	statement *ast.IfStmt,
) (verb string) {
	if slices.Contains(
		context.guards,
		ast.Unparen(statement.Cond),
	) {
		return noLogic
	}

	return "branches"
}

func computationVerb(
	context delegation,
	node ast.Node,
) (verb string) {
	switch typed := node.(type) {
	case *ast.BinaryExpr:
		return guardVerb(
			context,
			typed,
		)
	case *ast.UnaryExpr:
		return unaryVerb(typed)
	case *ast.AssignStmt:
		return assignmentVerb(
			context,
			typed,
		)
	case *ast.IncDecStmt:
		return "computes"
	}

	return noLogic
}

func guardVerb(
	context delegation,
	expression *ast.BinaryExpr,
) (verb string) {
	if slices.Contains(
		context.guards,
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
	context delegation,
	statement *ast.AssignStmt,
) (verb string) {
	isPlain := statement.Tok == token.ASSIGN || statement.Tok == token.DEFINE
	if !isPlain {
		return "computes"
	}

	fields := []string{}
	for _, target := range statement.Lhs {
		fields = append(
			fields,
			localField(
				context,
				target,
			)...,
		)
	}

	if len(fields) == 0 {
		return noLogic
	}

	return fieldVerb + fields[0] + " by hand"
}

func localField(
	context delegation,
	target ast.Expr,
) (fields []string) {
	selector, isSelector := ast.Unparen(target).(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}

	owner, isIdentifier := ast.Unparen(selector.X).(*ast.Ident)
	isLocal := isIdentifier && slices.Contains(
		context.locals,
		owner.Name,
	)
	if !isLocal {
		return nil
	}

	return []string{owner.Name + "." + selector.Sel.Name}
}

func localNames(
	function *ast.FuncDecl,
) (names []string) {
	names = slices.Concat(
		fieldNames(function.Type.Params),
		fieldNames(function.Type.Results),
	)
	for _, statement := range source.Nodes[*ast.AssignStmt](function.Body) {
		names = append(
			names,
			definedNames(statement)...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](function.Body) {
		names = append(
			names,
			identifierNames(spec.Names)...,
		)
	}

	return names
}

func definedNames(
	statement *ast.AssignStmt,
) (names []string) {
	if statement.Tok != token.DEFINE {
		return nil
	}

	for _, target := range statement.Lhs {
		names = append(
			names,
			expressionName(target)...,
		)
	}

	return names
}

func expressionName(
	expression ast.Expr,
) (names []string) {
	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	return []string{identifier.Name}
}

func fieldNames(
	fields *ast.FieldList,
) (names []string) {
	if fields == nil {
		return nil
	}

	for _, field := range fields.List {
		names = append(
			names,
			identifierNames(field.Names)...,
		)
	}

	return names
}

func identifierNames(
	identifiers []*ast.Ident,
) (names []string) {
	for _, identifier := range identifiers {
		names = append(
			names,
			identifier.Name,
		)
	}

	return names
}
