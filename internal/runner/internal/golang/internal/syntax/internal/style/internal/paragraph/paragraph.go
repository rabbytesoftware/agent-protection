package paragraph

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R55"
	afterBlock      = "statement right after a closed block: leave one blank line between the block's closing brace and this statement"
	detachedCheck   = "if checks a value the statement above it just assigned: remove the blank line so the check stays attached to that statement"
	leadingBlank    = "block starts with a blank line: remove it"
	trailingBlank   = "block ends with a blank line: remove it"
	minimumSequence = 2
)

func Paragraph(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	for _, statements := range sequences(file.Syntax) {
		findings = append(
			findings,
			sequenceFindings(
				file,
				statements,
			)...,
		)
	}

	for _, block := range source.Nodes[*ast.BlockStmt](file.Syntax) {
		findings = append(
			findings,
			edgeFindings(
				file,
				block,
			)...,
		)
	}

	return findings
}

func sequences(
	syntax *ast.File,
) (statementLists [][]ast.Stmt) {
	for _, block := range source.Nodes[*ast.BlockStmt](syntax) {
		statementLists = append(
			statementLists,
			block.List,
		)
	}

	for _, clause := range source.Nodes[*ast.CaseClause](syntax) {
		statementLists = append(
			statementLists,
			clause.Body,
		)
	}

	for _, clause := range source.Nodes[*ast.CommClause](syntax) {
		statementLists = append(
			statementLists,
			clause.Body,
		)
	}

	return slices.DeleteFunc(
		statementLists,
		isTooShort,
	)
}

func sequenceFindings(
	file source.File,
	statements []ast.Stmt,
) (findings []report.Finding) {
	for index := 1; index < len(statements); index++ {
		findings = append(
			findings,
			pairFindings(
				file,
				statements[index-1],
				statements[index],
			)...,
		)
	}

	return findings
}

func pairFindings(
	file source.File,
	previous ast.Stmt,
	current ast.Stmt,
) (findings []report.Finding) {
	snippet := pairViolation(
		file,
		previous,
		current,
	)
	if snippet == "" {
		return nil
	}

	return []report.Finding{finding(
		file,
		current,
		snippet,
	)}
}

func pairViolation(
	file source.File,
	previous ast.Stmt,
	current ast.Stmt,
) (snippet string) {
	separated := blankBetween(
		file,
		previous.End(),
		current.Pos(),
	)
	isCrampedBlock := closesBlock(previous) && !separated
	if isCrampedBlock {
		return afterBlock
	}

	isDetachedCheck := separated && checksAssignment(
		previous,
		current,
	)
	if isDetachedCheck {
		return detachedCheck
	}

	return ""
}

func edgeFindings(
	file source.File,
	block *ast.BlockStmt,
) (findings []report.Finding) {
	if len(block.List) == 0 {
		return nil
	}

	if blankBetween(
		file,
		block.Lbrace,
		block.List[0].Pos(),
	) {
		findings = append(
			findings,
			finding(
				file,
				block,
				leadingBlank,
			),
		)
	}

	if blankBetween(
		file,
		block.List[len(block.List)-1].End(),
		block.Rbrace,
	) {
		findings = append(
			findings,
			finding(
				file,
				block,
				trailingBlank,
			),
		)
	}

	return findings
}

func closesBlock(
	statement ast.Stmt,
) (closes bool) {
	switch typed := statement.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.BlockStmt:
		return true
	case *ast.LabeledStmt:
		return closesBlock(typed.Stmt)
	}

	return false
}

func checksAssignment(
	previous ast.Stmt,
	current ast.Stmt,
) (checks bool) {
	check, isIf := current.(*ast.IfStmt)
	if !isIf {
		return false
	}

	return check.Init == nil && usesAny(
		check.Cond,
		assignedNames(previous),
	)
}

func usesAny(
	expression ast.Expr,
	names []string,
) (uses bool) {
	for _, identifier := range source.Nodes[*ast.Ident](expression) {
		uses = uses || slices.Contains(
			names,
			identifier.Name,
		)
	}

	return uses
}

func assignedNames(
	statement ast.Stmt,
) (names []string) {
	assignment, isAssignment := statement.(*ast.AssignStmt)
	if !isAssignment {
		return nil
	}

	for _, target := range assignment.Lhs {
		names = append(
			names,
			identifierName(target),
		)
	}

	return names
}

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier {
		return ""
	}

	return identifier.Name
}

func finding(
	file source.File,
	node ast.Node,
	snippet string,
) (result report.Finding) {
	return report.Finding{
		Rule:     rule,
		Location: file.Location(node),
		Snippet:  snippet,
	}
}

func isTooShort(
	statements []ast.Stmt,
) (short bool) {
	return len(statements) < minimumSequence
}
