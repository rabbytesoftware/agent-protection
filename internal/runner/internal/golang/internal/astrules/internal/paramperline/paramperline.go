package paramperline

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R01"
	anonymous     = "func literal"
	snippetFormat = "%s: put every parameter on its own line, after an opening paren that ends the signature line and before a closing paren on a line of its own"
)

// Paramperline reports every function, method, func literal, interface
// method and func type whose parameters do not each sit on their own line
// between an opening paren ending the signature line and a closing paren on
// its own line. Receivers and results are not checked.
func Paramperline(
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
	for _, node := range source.Nodes[ast.Node](file.Syntax) {
		findings = append(
			findings,
			nodeFindings(
				file,
				node,
			)...,
		)
	}
	return findings
}

func nodeFindings(
	file source.File,
	node ast.Node,
) (findings []report.Finding) {
	name, funcType := signature(node)
	isCompliant := funcType == nil || hasParameterPerLine(
		file,
		funcType.Params,
	)
	if isCompliant {
		return nil
	}
	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(node),
		Snippet: fmt.Sprintf(
			snippetFormat,
			name,
		),
	}}
}

func signature(
	node ast.Node,
) (name string, funcType *ast.FuncType) {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		return typed.Name.Name,
			typed.Type
	case *ast.FuncLit:
		return anonymous,
			typed.Type
	case *ast.TypeSpec:
		return typeSignature(typed)
	case *ast.Field:
		return methodSignature(typed)
	}
	return "",
		nil
}

func typeSignature(
	spec *ast.TypeSpec,
) (name string, funcType *ast.FuncType) {
	funcType, isFunc := spec.Type.(*ast.FuncType)
	if !isFunc {
		return "",
			nil
	}
	return spec.Name.Name,
		funcType
}

func methodSignature(
	field *ast.Field,
) (name string, funcType *ast.FuncType) {
	funcType, isFunc := field.Type.(*ast.FuncType)
	isInterfaceMethod := isFunc && !funcType.Func.IsValid() && len(field.Names) > 0
	if !isInterfaceMethod {
		return "",
			nil
	}
	return field.Names[0].Name,
		funcType
}

func hasParameterPerLine(
	file source.File,
	params *ast.FieldList,
) (compliant bool) {
	if len(params.List) == 0 {
		return true
	}
	opening := span{first: lineOf(
		file,
		params.Opening,
	), last: lineOf(
		file,
		params.Opening,
	)}
	closing := span{first: lineOf(
		file,
		params.Closing,
	), last: lineOf(
		file,
		params.Closing,
	)}
	spans := slices.Concat(
		[]span{opening},
		parameterSpans(
			file,
			params,
		),
		[]span{closing},
	)
	gaps := []span{}
	for index := 1; index < len(spans); index++ {
		gaps = append(
			gaps,
			span{first: spans[index-1].last, last: spans[index].first},
		)
	}
	return !slices.ContainsFunc(
		gaps,
		isCollapsed,
	)
}

func isCollapsed(
	gap span,
) (collapsed bool) {
	return gap.last <= gap.first
}

func parameterSpans(
	file source.File,
	params *ast.FieldList,
) (spans []span) {
	for _, field := range params.List {
		spans = append(
			spans,
			fieldSpans(
				file,
				field,
			)...,
		)
	}
	return spans
}

func fieldSpans(
	file source.File,
	field *ast.Field,
) (spans []span) {
	if len(field.Names) == 0 {
		return []span{{first: lineOf(
			file,
			field.Pos(),
		), last: lineOf(
			file,
			field.End(),
		)}}
	}
	for _, name := range field.Names {
		spans = append(
			spans,
			span{first: lineOf(
				file,
				name.Pos(),
			), last: lineOf(
				file,
				name.End(),
			)},
		)
	}
	spans[len(spans)-1].last = lineOf(
		file,
		field.End(),
	)
	return spans
}

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
