package concretetype

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const writerDetail = "value-receiver method %s that writes through its field %s"

func writtenThrough(
	files []source.File,
) (writers map[string]string) {
	writers = map[string]string{}
	fields := referenceFields(files)
	declared := packageMethods(files)
	errorTypes := errorTypeNames(declared)
	for _, method := range slices.DeleteFunc(
		declared,
		isPointerReceiver,
	) {
		recordWriter(
			writers,
			method,
			fields,
			errorTypes,
		)
	}

	return writers
}

func recordWriter(
	writers map[string]string,
	method *ast.FuncDecl,
	fields map[string][]string,
	errorTypes []string,
) {
	name, _ := receiver(method.Recv.List[0].Type)
	field := writtenField(
		method,
		fields[name],
	)
	isWriter := field != "" && !slices.Contains(
		errorTypes,
		name,
	)
	if !isWriter {
		return
	}

	writers[name] = fmt.Sprintf(
		writerDetail,
		method.Name.Name,
		field,
	)
}

func isPointerReceiver(
	method *ast.FuncDecl,
) (excluded bool) {
	_, isPointer := receiver(method.Recv.List[0].Type)
	return isPointer
}

func referenceFields(
	files []source.File,
) (fields map[string][]string) {
	fields = map[string][]string{}
	for _, file := range files {
		recordReferences(
			fields,
			file,
		)
	}

	return fields
}

func recordReferences(
	fields map[string][]string,
	file source.File,
) {
	for _, spec := range source.Nodes[*ast.TypeSpec](file.Syntax) {
		fields[spec.Name.Name] = append(
			fields[spec.Name.Name],
			structReferences(spec.Type)...,
		)
	}
}

func structReferences(
	expression ast.Expr,
) (names []string) {
	structure, isStruct := expression.(*ast.StructType)
	if !isStruct {
		return nil
	}

	for _, field := range structure.Fields.List {
		names = appendReferences(
			names,
			field,
		)
	}

	return names
}

func appendReferences(
	names []string,
	field *ast.Field,
) (appended []string) {
	if !isReference(field.Type) {
		return names
	}

	if len(field.Names) == 0 {
		embedded, _ := receiver(field.Type)
		return append(
			names,
			embedded,
		)
	}

	for _, name := range field.Names {
		names = append(
			names,
			name.Name,
		)
	}

	return names
}

func isReference(
	expression ast.Expr,
) (reference bool) {
	switch typed := expression.(type) {
	case *ast.MapType, *ast.ChanType, *ast.StarExpr:
		return true
	case *ast.ArrayType:
		return typed.Len == nil
	}

	return false
}

func writtenField(
	method *ast.FuncDecl,
	fields []string,
) (field string) {
	names := method.Recv.List[0].Names
	isUsable := method.Body != nil && len(names) == 1 && len(fields) > 0
	if !isUsable {
		return ""
	}

	for _, target := range throughTargets(method.Body) {
		field = cmp.Or(
			field,
			throughField(
				target,
				names[0].Name,
				fields,
			),
		)
	}

	for _, target := range directTargets(method.Body) {
		field = cmp.Or(
			field,
			fieldOf(
				target,
				names[0].Name,
				fields,
			),
		)
	}

	return field
}

func throughTargets(
	body *ast.BlockStmt,
) (targets []ast.Expr) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](body) {
		targets = appendAssigned(
			targets,
			assignment,
		)
	}

	for _, statement := range source.Nodes[*ast.IncDecStmt](body) {
		targets = append(
			targets,
			statement.X,
		)
	}

	return targets
}

func appendAssigned(
	targets []ast.Expr,
	assignment *ast.AssignStmt,
) (appended []ast.Expr) {
	if assignment.Tok == token.DEFINE {
		return targets
	}

	return append(
		targets,
		assignment.Lhs...,
	)
}

func directTargets(
	body *ast.BlockStmt,
) (targets []ast.Expr) {
	for _, statement := range source.Nodes[*ast.SendStmt](body) {
		targets = append(
			targets,
			statement.Chan,
		)
	}

	for _, call := range source.Nodes[*ast.CallExpr](body) {
		targets = appendCleared(
			targets,
			call,
		)
	}

	return targets
}

func appendCleared(
	targets []ast.Expr,
	call *ast.CallExpr,
) (appended []ast.Expr) {
	name, isIdent := call.Fun.(*ast.Ident)
	isClearing := isIdent &&
		name.Obj == nil &&
		len(call.Args) > 0 &&
		slices.Contains(
			clearingBuiltins(),
			name.Name,
		)
	if !isClearing {
		return targets
	}

	return append(
		targets,
		call.Args[0],
	)
}

func clearingBuiltins() (names []string) {
	return []string{
		"delete",
		"clear",
	}
}

func throughField(
	expression ast.Expr,
	receiverName string,
	fields []string,
) (field string) {
	var inner ast.Expr
	switch typed := ast.Unparen(expression).(type) {
	case *ast.IndexExpr:
		inner = typed.X
	case *ast.StarExpr:
		inner = typed.X
	case *ast.SelectorExpr:
		inner = typed.X
	default:
		return ""
	}

	return cmp.Or(
		fieldOf(
			inner,
			receiverName,
			fields,
		),
		throughField(
			inner,
			receiverName,
			fields,
		),
	)
}

func fieldOf(
	expression ast.Expr,
	receiverName string,
	fields []string,
) (field string) {
	selector, isSelector := ast.Unparen(expression).(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}

	owner, isIdent := ast.Unparen(selector.X).(*ast.Ident)
	isField := isIdent && owner.Name == receiverName && slices.Contains(
		fields,
		selector.Sel.Name,
	)
	if !isField {
		return ""
	}

	return selector.Sel.Name
}
