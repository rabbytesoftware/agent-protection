package concretetype

import (
	"fmt"
	"go/ast"
	"go/types"
	"maps"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule        = "R13"
	mainPackage = "main"
	errorMethod = "Error"
)

func Concretetype(
	files []source.File,
) (findings []report.Finding) {
	packages := map[string][]source.File{}
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isExempt,
	) {
		directory := filepath.Dir(file.Rel)
		packages[directory] = append(
			packages[directory],
			file,
		)
	}

	for _, directory := range slices.Sorted(maps.Keys(packages)) {
		findings = append(
			findings,
			packageFindings(packages[directory])...,
		)
	}

	return findings
}

func packageFindings(
	files []source.File,
) (findings []report.Finding) {
	implementations := implementationNames(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				file,
				implementations,
			)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
	implementations []string,
) (findings []report.Finding) {
	for _, spec := range exportedStructs(file) {
		findings = append(
			findings,
			specFindings(
				file,
				spec,
				implementations,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	spec *ast.TypeSpec,
	implementations []string,
) (findings []report.Finding) {
	if !slices.Contains(
		implementations,
		spec.Name.Name,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"exported struct %s has pointer-receiver methods: it is an implementation; unexport it and return an interface from its constructor",
			spec.Name.Name,
		),
	}}
}

func implementationNames(
	files []source.File,
) (names []string) {
	declared := packageMethods(files)
	errorTypes := receiverNames(slices.DeleteFunc(
		slices.Clone(declared),
		isNotErrorMethod,
	))
	for _, name := range receiverNames(slices.DeleteFunc(
		declared,
		isValueReceiver,
	)) {
		names = appendImplementation(
			names,
			name,
			errorTypes,
		)
	}

	return names
}

func appendImplementation(
	names []string,
	name string,
	errorTypes []string,
) (appended []string) {
	if slices.Contains(
		errorTypes,
		name,
	) {
		return names
	}

	return append(
		names,
		name,
	)
}

func packageMethods(
	files []source.File,
) (found []*ast.FuncDecl) {
	for _, file := range files {
		found = append(
			found,
			methods(file)...,
		)
	}

	return found
}

func receiverNames(
	declared []*ast.FuncDecl,
) (names []string) {
	for _, method := range declared {
		name, _ := receiver(method.Recv.List[0].Type)
		names = append(
			names,
			name,
		)
	}

	return names
}

func isNotErrorMethod(
	method *ast.FuncDecl,
) (excluded bool) {
	return method.Name.Name != errorMethod
}

func isValueReceiver(
	method *ast.FuncDecl,
) (excluded bool) {
	_, isPointer := receiver(method.Recv.List[0].Type)
	return !isPointer
}

func receiver(
	expression ast.Expr,
) (name string, isPointer bool) {
	switch typed := expression.(type) {
	case *ast.StarExpr:
		name, _ = receiver(typed.X)
		return name,
			true
	case *ast.ParenExpr:
		return receiver(typed.X)
	case *ast.IndexExpr:
		return receiver(typed.X)
	case *ast.IndexListExpr:
		return receiver(typed.X)
	}

	return types.ExprString(expression),
		false
}

func methods(
	file source.File,
) (found []*ast.FuncDecl) {
	return slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotMethod,
	)
}

func exportedStructs(
	file source.File,
) (specs []*ast.TypeSpec) {
	return slices.DeleteFunc(
		source.Nodes[*ast.TypeSpec](file.Syntax),
		isNotExportedStruct,
	)
}

func isNotExportedStruct(
	spec *ast.TypeSpec,
) (excluded bool) {
	_, isStruct := spec.Type.(*ast.StructType)
	return !isStruct || spec.Assign.IsValid() || !spec.Name.IsExported()
}

func isNotMethod(
	function *ast.FuncDecl,
) (notMethod bool) {
	return function.Recv == nil || len(function.Recv.List) == 0
}

func isExempt(
	file source.File,
) (exempt bool) {
	return source.IsTest(file.Path) || source.IsMock(file.Rel) || file.Syntax.Name.Name == mainPackage
}
