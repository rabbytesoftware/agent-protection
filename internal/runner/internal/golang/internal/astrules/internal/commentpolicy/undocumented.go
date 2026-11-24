package commentpolicy

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func undocumentedFindings(
	file source.File,
	decl ast.Decl,
) (findings []report.Finding) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return undocumentedFunc(
			file,
			typed,
		)
	case *ast.GenDecl:
		return undocumentedSpecs(
			file,
			typed,
		)
	}
	return nil
}

func undocumentedFunc(
	file source.File,
	decl *ast.FuncDecl,
) (findings []report.Finding) {
	isMissing := needsDoc(
		file,
		decl,
	) && !hasDoc(decl.Doc)
	if !isMissing {
		return nil
	}
	return missingDoc(
		file,
		decl,
		decl.Name.Name,
	)
}

func undocumentedSpecs(
	file source.File,
	decl *ast.GenDecl,
) (findings []report.Finding) {
	for _, spec := range decl.Specs {
		findings = append(
			findings,
			undocumentedSpec(
				file,
				decl,
				spec,
			)...,
		)
	}
	return findings
}

func undocumentedSpec(
	file source.File,
	decl *ast.GenDecl,
	spec ast.Spec,
) (findings []report.Finding) {
	name := exportedName(spec)
	isMissing := name != "" && !hasDoc(specDoc(spec)) && !hasDoc(decl.Doc)
	if !isMissing {
		return nil
	}
	return missingDoc(
		file,
		spec,
		name,
	)
}

func missingDoc(
	file source.File,
	node ast.Node,
	name string,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(node),
		Snippet: fmt.Sprintf(
			missingDocSnippet,
			name,
		),
	}}
}

func needsDoc(
	file source.File,
	decl *ast.FuncDecl,
) (needed bool) {
	return decl.Name.IsExported() && !isTestFunction(
		file,
		decl,
	)
}

func isTestFunction(
	file source.File,
	decl *ast.FuncDecl,
) (isTest bool) {
	isTopLevelTestFile := source.IsTest(file.Path) && decl.Recv == nil
	if !isTopLevelTestFile {
		return false
	}
	name := decl.Name.Name
	isExample := strings.HasPrefix(
		name,
		examplePrefix,
	) && decl.Type.Params.NumFields() == 0
	hasTestPrefix := strings.HasPrefix(
		name,
		testPrefix,
	) || strings.HasPrefix(
		name,
		benchmarkPrefix,
	) || strings.HasPrefix(
		name,
		fuzzPrefix,
	)
	isTestRunner := hasTestPrefix && takesTestingPointer(decl.Type.Params)
	return isExample || isTestRunner
}

func takesTestingPointer(
	params *ast.FieldList,
) (takes bool) {
	if params.NumFields() != 1 {
		return false
	}
	pointer, isPointer := params.List[0].Type.(*ast.StarExpr)
	if !isPointer {
		return false
	}
	selector, isSelector := pointer.X.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}
	pkg, isIdent := selector.X.(*ast.Ident)
	return isIdent && pkg.Name == testingPackage
}
