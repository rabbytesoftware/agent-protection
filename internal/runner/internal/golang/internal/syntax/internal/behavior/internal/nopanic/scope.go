package nopanic

import (
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const loggerSuffix = "Logger"

type scope struct {
	imports map[string]string
	fields  []string
	callees []ast.Expr
}

func scopeOf(
	syntax *ast.File,
	fields []string,
) (found scope) {
	found = scope{
		imports: importedPaths(syntax),
		fields:  fields,
	}
	for _, call := range source.Nodes[*ast.CallExpr](syntax) {
		found.callees = append(
			found.callees,
			call.Fun,
		)
	}

	return found
}

func (s scope) isCalled(
	selector *ast.SelectorExpr,
) (called bool) {
	return slices.Contains(
		s.callees,
		ast.Expr(selector),
	)
}

func (s scope) isLogger(
	expression ast.Expr,
) (logger bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.CallExpr:
		return s.isLoggerConstruction(typed)
	case *ast.Ident:
		return s.isLoggerVariable(typed)
	case *ast.SelectorExpr:
		return slices.Contains(
			s.fields,
			typed.Sel.Name,
		)
	}

	return false
}

func (s scope) isLoggerConstruction(
	call *ast.CallExpr,
) (constructs bool) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	receiver, isIdent := selector.X.(*ast.Ident)
	return isIdent && receiver.Obj == nil && s.imports[receiver.Name] == logPath && slices.Contains(
		loggerConstructors(),
		selector.Sel.Name,
	)
}

func (s scope) isLoggerVariable(
	name *ast.Ident,
) (logger bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.Field:
		return s.isLoggerType(declaration.Type)
	case *ast.ValueSpec:
		return s.isLoggerType(declaration.Type) || s.isBoundToLogger(declaration.Values)
	case *ast.AssignStmt:
		return s.isBoundToLogger(declaration.Rhs)
	}

	return false
}

func (s scope) isBoundToLogger(
	values []ast.Expr,
) (bound bool) {
	for _, value := range values {
		bound = bound || s.isLogger(value)
	}

	return bound
}

func (s scope) isLoggerType(
	expression ast.Expr,
) (logger bool) {
	switch typed := expression.(type) {
	case *ast.StarExpr:
		return s.isLoggerType(typed.X)
	case *ast.Ident:
		return strings.HasSuffix(
			typed.Name,
			loggerSuffix,
		)
	case *ast.SelectorExpr:
		_, isImported := s.imports[identifierName(typed.X)]
		return isImported && strings.HasSuffix(
			typed.Sel.Name,
			loggerSuffix,
		)
	}

	return false
}

func loggerFields(
	files []source.File,
) (fields map[string][]string) {
	fields = map[string][]string{}
	for _, file := range files {
		directory := filepath.Dir(file.Path)
		fields[directory] = append(
			fields[directory],
			fileLoggerFields(file.Syntax)...,
		)
	}

	return fields
}

func fileLoggerFields(
	syntax *ast.File,
) (names []string) {
	found := scope{imports: importedPaths(syntax)}
	for _, structure := range source.Nodes[*ast.StructType](syntax) {
		names = append(
			names,
			found.loggerFieldNames(structure)...,
		)
	}

	return names
}

func (s scope) loggerFieldNames(
	structure *ast.StructType,
) (names []string) {
	for _, field := range structure.Fields.List {
		names = s.appendLoggerField(
			names,
			field,
		)
	}

	return names
}

func (s scope) appendLoggerField(
	names []string,
	field *ast.Field,
) (extended []string) {
	if !s.isLoggerType(field.Type) {
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

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return identifier.Name
}

func loggerConstructors() (names []string) {
	return []string{
		"New",
		"Default",
	}
}
