package boolparam

import (
	"cmp"
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const fieldFormat = "func %s(...) branches on %s.%s, a bool field of its %s parameter: a flag moved into a struct is still a flag; split it into two named functions or make the field a typed enum with named constants"

func flagFieldFindings(
	file source.File,
	members []source.File,
) (findings []report.Finding) {
	fields := boolFields(members)
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFlagFindings(
				file,
				members,
				function,
				fields,
			)...,
		)
	}

	return findings
}

func boolFields(
	members []source.File,
) (fields map[string][]string) {
	boolNames := boolTypes(members)
	fields = map[string][]string{}
	specs := []*ast.TypeSpec{}
	for _, file := range members {
		specs = append(
			specs,
			source.Nodes[*ast.TypeSpec](file.Syntax)...,
		)
	}

	for _, spec := range specs {
		fields[spec.Name.Name] = structBools(
			spec,
			boolNames,
		)
	}

	return fields
}

func structBools(
	spec *ast.TypeSpec,
	boolNames []string,
) (names []string) {
	structure, isStruct := spec.Type.(*ast.StructType)
	if !isStruct {
		return nil
	}

	for _, field := range structure.Fields.List {
		names = appendBoolField(
			names,
			field,
			boolNames,
		)
	}

	return names
}

func appendBoolField(
	names []string,
	field *ast.Field,
	boolNames []string,
) (extended []string) {
	if !slices.Contains(
		boolNames,
		identifierName(field.Type),
	) {
		return names
	}

	for _, name := range field.Names {
		names = append(
			names,
			name.Name,
		)
	}

	return names
}

func functionFlagFindings(
	file source.File,
	members []source.File,
	function *ast.FuncDecl,
	fields map[string][]string,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	conditions := branchConditions(function.Body)
	names, kinds := parameterList(function.Type.Params)
	for position, name := range names {
		structName := identifierName(dereferenced(kinds[position]))
		findings = appendFieldFinding(
			findings,
			file,
			function,
			fieldFormat,
			name.Name,
			builtFlag(
				members,
				function.Name.Name,
				position,
				structName,
				branchedField(
					conditions,
					name.Name,
					fields[structName],
				),
			),
			structName,
		)
	}

	return findings
}

func builtFlag(
	members []source.File,
	function string,
	position int,
	structName string,
	flag string,
) (built string) {
	if flag == notFoundName {
		return notFoundName
	}

	if !builtAtCall(
		members,
		function,
		position,
		structName,
	) {
		return notFoundName
	}

	return flag
}

func appendFieldFinding(
	findings []report.Finding,
	file source.File,
	function *ast.FuncDecl,
	format string,
	parameter string,
	flag string,
	structName string,
) (extended []report.Finding) {
	if flag == notFoundName {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(function),
			Snippet: fmt.Sprintf(
				format,
				function.Name.Name,
				parameter,
				flag,
				structName,
			),
		},
	)
}

func dereferenced(
	expression ast.Expr,
) (base ast.Expr) {
	pointer, isPointer := expression.(*ast.StarExpr)
	if !isPointer {
		return expression
	}

	return pointer.X
}

func branchedField(
	conditions []ast.Expr,
	parameter string,
	flags []string,
) (flag string) {
	for _, condition := range conditions {
		flag = firstFlag(
			flag,
			source.Nodes[*ast.SelectorExpr](condition),
			parameter,
			flags,
		)
	}

	return flag
}

func firstFlag(
	found string,
	selectors []*ast.SelectorExpr,
	parameter string,
	flags []string,
) (flag string) {
	for _, selector := range selectors {
		found = cmp.Or(
			found,
			selectedFlag(
				selector,
				parameter,
				flags,
			),
		)
	}

	return found
}

func selectedFlag(
	selector *ast.SelectorExpr,
	parameter string,
	flags []string,
) (flag string) {
	isFlag := identifierName(selector.X) == parameter && slices.Contains(
		flags,
		selector.Sel.Name,
	)
	if !isFlag {
		return notFoundName
	}

	return selector.Sel.Name
}

func branchConditions(
	body *ast.BlockStmt,
) (conditions []ast.Expr) {
	for _, statement := range source.Nodes[*ast.IfStmt](body) {
		conditions = append(
			conditions,
			statement.Cond,
		)
	}

	for _, statement := range source.Nodes[*ast.SwitchStmt](body) {
		conditions = append(
			conditions,
			switchConditions(statement)...,
		)
	}

	return conditions
}

func switchConditions(
	statement *ast.SwitchStmt,
) (conditions []ast.Expr) {
	if statement.Tag != nil {
		return []ast.Expr{statement.Tag}
	}

	for _, clause := range source.Nodes[*ast.CaseClause](statement.Body) {
		conditions = append(
			conditions,
			clause.List...,
		)
	}

	return conditions
}
