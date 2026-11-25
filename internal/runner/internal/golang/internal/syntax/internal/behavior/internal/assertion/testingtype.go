package assertion

import (
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	testingPath  = "testing"
	suiteTesting = "T"
	testType     = "T"
	fuzzType     = "F"
	testPrefix   = "Test"
	fuzzPrefix   = "Fuzz"
)

func testingAlias(
	imports map[string]string,
) (alias string) {
	for name, importPath := range imports {
		alias = chosenAlias(
			alias,
			name,
			importPath,
		)
	}

	if alias == "" {
		return testingPath
	}

	return alias
}

func chosenAlias(
	current string,
	name string,
	importPath string,
) (alias string) {
	if importPath != testingPath {
		return current
	}

	return name
}

func isTestingType(
	expression ast.Expr,
	alias string,
	helpers []string,
) (testing bool) {
	switch typed := expression.(type) {
	case *ast.StarExpr:
		return isTestingSelector(
			typed.X,
			alias,
			pointerTypes(),
		)
	case *ast.SelectorExpr:
		return isTestingSelector(
			typed,
			alias,
			interfaceTypes(),
		) || slices.Contains(
			helpers,
			typed.Sel.Name,
		)
	case *ast.Ident:
		return slices.Contains(
			helpers,
			typed.Name,
		)
	}

	return false
}

func isTestingSelector(
	expression ast.Expr,
	alias string,
	names []string,
) (testing bool) {
	selector, isSelector := expression.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	return isIdent && qualifier.Name == alias && slices.Contains(
		names,
		selector.Sel.Name,
	)
}

func testingNames(
	owner declaration,
	helpers []string,
) (names []string) {
	alias := testingAlias(owner.imports)
	functions := append(
		[]*ast.FuncType{owner.function},
		source.Nodes[*ast.FuncType](owner.body)...,
	)
	for _, function := range functions {
		names = append(
			names,
			fieldNames(
				function.Params,
				alias,
				helpers,
			)...,
		)
	}

	return names
}

func fieldNames(
	fields *ast.FieldList,
	alias string,
	helpers []string,
) (names []string) {
	if fields == nil {
		return nil
	}

	for _, field := range fields.List {
		names = appendTestingField(
			names,
			field,
			alias,
			helpers,
		)
	}

	return names
}

func appendTestingField(
	names []string,
	field *ast.Field,
	alias string,
	helpers []string,
) (extended []string) {
	if !isTestingType(
		field.Type,
		alias,
		helpers,
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

func failingInterfaces(
	files []source.File,
) (names []string) {
	for _, file := range files {
		names = fileFailingInterfaces(
			names,
			file,
		)
	}

	return names
}

func fileFailingInterfaces(
	names []string,
	file source.File,
) (extended []string) {
	for _, spec := range source.Nodes[*ast.TypeSpec](file.Syntax) {
		names = appendFailingInterface(
			names,
			spec,
		)
	}

	return names
}

func appendFailingInterface(
	names []string,
	spec *ast.TypeSpec,
) (extended []string) {
	contract, isInterface := spec.Type.(*ast.InterfaceType)
	if !isInterface {
		return names
	}

	for _, method := range contract.Methods.List {
		names = appendFailingMethod(
			names,
			spec,
			method,
		)
	}

	return names
}

func appendFailingMethod(
	names []string,
	spec *ast.TypeSpec,
	method *ast.Field,
) (extended []string) {
	for _, name := range method.Names {
		names = appendIfFailing(
			names,
			spec.Name.Name,
			name.Name,
		)
	}

	return names
}

func appendIfFailing(
	names []string,
	typeName string,
	methodName string,
) (extended []string) {
	isFailing := slices.Contains(
		failingMethods(),
		methodName,
	) && !slices.Contains(
		names,
		typeName,
	)
	if !isFailing {
		return names
	}

	return append(
		names,
		typeName,
	)
}

func testingFields(
	files []source.File,
	helpers []string,
) (fields map[string][]string) {
	fields = map[string][]string{}
	for _, file := range files {
		directory := filepath.Dir(file.Rel)
		fields[directory] = append(
			fields[directory],
			structFields(
				file,
				helpers,
			)...,
		)
	}

	return fields
}

func structFields(
	file source.File,
	helpers []string,
) (names []string) {
	alias := testingAlias(importedPaths(file.Syntax))
	for _, structure := range source.Nodes[*ast.StructType](file.Syntax) {
		names = append(
			names,
			fieldNames(
				structure.Fields,
				alias,
				helpers,
			)...,
		)
	}

	return names
}

func pointerTypes() (names []string) {
	return []string{
		"T",
		"B",
		"F",
	}
}

func interfaceTypes() (names []string) {
	return []string{"TB"}
}

func failingMethods() (names []string) {
	return []string{
		"Error",
		"Errorf",
		"Fatal",
		"Fatalf",
		"Fail",
		"FailNow",
	}
}
