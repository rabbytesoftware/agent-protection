package publicapi

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const aliasFormat = "type alias %s exposes %s: declare an interface for what callers need"

func aliasFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, spec := range slices.DeleteFunc(
		typeSpecs(declaration),
		isNotExportedAlias,
	) {
		findings = append(
			findings,
			exposingAlias(
				file,
				found,
				spec,
			)...,
		)
	}

	return findings
}

func exposingAlias(
	file source.File,
	found scope,
	spec *ast.TypeSpec,
) (findings []report.Finding) {
	target := aliasTarget(spec.Type)
	isExposing := isStructTarget(
		target,
		found.structs,
	) || isInternalTarget(
		file.Syntax,
		target,
		found.vocabularies,
	)
	if !isExposing {
		return nil
	}

	return []report.Finding{{
		Rule:     apiRule,
		Location: file.Location(spec.Name),
		Snippet: fmt.Sprintf(
			aliasFormat,
			spec.Name.Name,
			types.ExprString(spec.Type),
		),
	}}
}

func aliasTarget(
	expression ast.Expr,
) (target ast.Expr) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.StarExpr:
		return aliasTarget(typed.X)
	case *ast.IndexExpr:
		return aliasTarget(typed.X)
	case *ast.IndexListExpr:
		return aliasTarget(typed.X)
	}

	return ast.Unparen(expression)
}

func isStructTarget(
	target ast.Expr,
	structs []string,
) (isStruct bool) {
	switch typed := target.(type) {
	case *ast.StructType:
		return true
	case *ast.Ident:
		return slices.Contains(
			structs,
			typed.Name,
		)
	}

	return false
}

func isInternalTarget(
	syntax *ast.File,
	target ast.Expr,
	vocabularies []string,
) (isInternal bool) {
	selector, isSelector := target.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdentifier := selector.X.(*ast.Ident)
	return isIdentifier && len(without(
		slices.DeleteFunc(
			qualifiedPaths(
				syntax,
				qualifier.Name,
			),
			isOutsideInternal,
		),
		vocabularies,
	)) > 0
}

func isOutsideInternal(
	imported string,
) (excluded bool) {
	return !source.IsInternalTree(imported)
}

func vocabularyImports(
	modulePath string,
	files []source.File,
) (paths []string) {
	for _, directory := range source.DeclarationPackages(files) {
		paths = append(
			paths,
			path.Join(
				modulePath,
				directory,
			),
		)
	}

	return paths
}

func qualifiedPaths(
	syntax *ast.File,
	qualifier string,
) (paths []string) {
	for _, spec := range syntax.Imports {
		paths = append(
			paths,
			pathNamed(
				spec,
				qualifier,
			)...,
		)
	}

	return paths
}

func pathNamed(
	spec *ast.ImportSpec,
	qualifier string,
) (paths []string) {
	imported := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	if importName(
		spec,
		imported,
	) != qualifier {
		return nil
	}

	return []string{imported}
}

func importName(
	spec *ast.ImportSpec,
	imported string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(imported)
}

func typeSpecs(
	declaration ast.Decl,
) (specs []*ast.TypeSpec) {
	general, isGeneral := declaration.(*ast.GenDecl)
	isTypeDeclaration := isGeneral && general.Tok == token.TYPE
	if !isTypeDeclaration {
		return nil
	}

	return source.Nodes[*ast.TypeSpec](general)
}

func packageStructs(
	files []source.File,
) (structs map[string][]string) {
	structs = map[string][]string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		structs[directory] = append(
			structs[directory],
			structNames(file.Syntax)...,
		)
	}

	return structs
}

func structNames(
	syntax *ast.File,
) (names []string) {
	for _, declaration := range syntax.Decls {
		names = append(
			names,
			declaredStructs(declaration)...,
		)
	}

	return names
}

func declaredStructs(
	declaration ast.Decl,
) (names []string) {
	for _, spec := range slices.DeleteFunc(
		typeSpecs(declaration),
		isNotStructDefinition,
	) {
		names = append(
			names,
			spec.Name.Name,
		)
	}

	return names
}

func isNotExportedAlias(
	spec *ast.TypeSpec,
) (excluded bool) {
	return !spec.Assign.IsValid() || !spec.Name.IsExported()
}

func isNotStructDefinition(
	spec *ast.TypeSpec,
) (excluded bool) {
	_, isStruct := spec.Type.(*ast.StructType)
	return spec.Assign.IsValid() || !isStruct
}
