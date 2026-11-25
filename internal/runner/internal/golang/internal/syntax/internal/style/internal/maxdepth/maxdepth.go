package maxdepth

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R02"
	deepestLevel  = 2
	snippetFormat = "%s: block nested %d levels deep, max %d: extract it into a named helper or flatten it with guard clauses"
	closureFormat = "%s: block nested %d levels deep where its local closure is used, max %d: hoisting a closure into a variable does not flatten it, extract a named function or method instead"
)

func Maxdepth(
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
	for _, decl := range file.Syntax.Decls {
		findings = append(
			findings,
			declFindings(
				file,
				decl,
			)...,
		)
	}

	return findings
}

func declFindings(
	file source.File,
	decl ast.Decl,
) (findings []report.Finding) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return funcFindings(
			file,
			typed,
		)
	case *ast.GenDecl:
		return genDeclFindings(
			file,
			typed,
		)
	}

	return nil
}

func funcFindings(
	file source.File,
	decl *ast.FuncDecl,
) (findings []report.Finding) {
	if decl.Body == nil {
		return nil
	}

	return blockFindings(
		file,
		decl.Name.Name,
		decl.Body,
		file.Column(decl),
	)
}

func genDeclFindings(
	file source.File,
	decl *ast.GenDecl,
) (findings []report.Finding) {
	for _, spec := range decl.Specs {
		findings = append(
			findings,
			specFindings(
				file,
				decl,
				spec,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	decl *ast.GenDecl,
	spec ast.Spec,
) (findings []report.Finding) {
	value, isValue := spec.(*ast.ValueSpec)
	if !isValue {
		return nil
	}

	if decl.Lparen.IsValid() {
		return blockFindings(
			file,
			value.Names[0].Name,
			value,
			file.Column(value),
		)
	}

	return blockFindings(
		file,
		value.Names[0].Name,
		value,
		file.Column(decl),
	)
}

func blockFindings(
	file source.File,
	owner string,
	root ast.Node,
	baseColumn int,
) (findings []report.Finding) {
	shifts := closureShifts(
		file,
		root,
	)
	for _, node := range source.Nodes[ast.Node](root) {
		findings = append(
			findings,
			listFindings(
				file,
				owner,
				node,
				baseColumn,
				shifts,
			)...,
		)
	}

	return findings
}

func listFindings(
	file source.File,
	owner string,
	node ast.Node,
	baseColumn int,
	shifts map[*ast.FuncLit]int,
) (findings []report.Finding) {
	opening, statements := statementList(node)
	if len(statements) == 0 {
		return nil
	}

	first := unlabeled(statements[0])
	shift := closureShift(
		node,
		shifts,
	)
	depth := file.Column(first) - baseColumn + shift
	isTooDeep := depth > deepestLevel && file.Line(first) != file.FileSet.Position(opening).Line
	if !isTooDeep {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(first),
		Snippet: snippet(
			owner,
			depth,
			shift,
		),
	}}
}

func snippet(
	owner string,
	depth int,
	shift int,
) (text string) {
	if shift > 0 {
		return fmt.Sprintf(
			closureFormat,
			owner,
			depth,
			deepestLevel,
		)
	}

	return fmt.Sprintf(
		snippetFormat,
		owner,
		depth,
		deepestLevel,
	)
}

func statementList(
	node ast.Node,
) (opening token.Pos, statements []ast.Stmt) {
	switch typed := node.(type) {
	case *ast.BlockStmt:
		return typed.Lbrace,
			slices.DeleteFunc(
				slices.Clone(typed.List),
				isClause,
			)
	case *ast.CaseClause:
		return typed.Colon,
			typed.Body
	case *ast.CommClause:
		return typed.Colon,
			typed.Body
	}

	return token.NoPos,
		nil
}

func isClause(
	statement ast.Stmt,
) (clause bool) {
	switch statement.(type) {
	case *ast.CaseClause, *ast.CommClause:
		return true
	}

	return false
}

func unlabeled(
	statement ast.Stmt,
) (inner ast.Stmt) {
	labeled, isLabeled := statement.(*ast.LabeledStmt)
	if !isLabeled {
		return statement
	}

	return unlabeled(labeled.Stmt)
}
