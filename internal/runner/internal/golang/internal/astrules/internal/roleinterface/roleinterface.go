package roleinterface

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule             = "R60"
	pathSeparator    = "/"
	currentDirectory = "."
)

func Roleinterface(
	files []source.File,
) (findings []report.Finding) {
	embedded, plain := usages(files)
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isExempt,
	) {
		findings = append(
			findings,
			fileFindings(
				file,
				embedded,
				plain,
			)...,
		)
	}

	return findings
}

func usages(
	files []source.File,
) (embedded map[string]bool, plain map[string]bool) {
	packageNames := declaredPackageNames(files)
	modulePath := derivedModulePath(files)
	references := []reference{}
	for _, file := range files {
		references = append(
			references,
			fileReferences(
				file,
				fileAliases(
					file,
					modulePath,
					packageNames,
				),
				packageNames,
			)...,
		)
	}

	embedded = map[string]bool{}
	plain = map[string]bool{}
	for _, used := range references {
		embedded[used.key] = embedded[used.key] || used.embedded
		plain[used.key] = plain[used.key] || !used.embedded
	}

	return embedded,
		plain
}

func fileFindings(
	file source.File,
	embedded map[string]bool,
	plain map[string]bool,
) (findings []report.Finding) {
	for _, spec := range interfaceSpecs(file) {
		findings = append(
			findings,
			specFindings(
				file,
				spec,
				embedded,
				plain,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	spec *ast.TypeSpec,
	embedded map[string]bool,
	plain map[string]bool,
) (findings []report.Finding) {
	reembedded, isReembedding := soleEmbedded(spec)
	if isReembedding {
		return []report.Finding{{
			Rule:     rule,
			Location: file.Location(spec),
			Snippet: fmt.Sprintf(
				"interface %s only re-embeds %s: use %s directly",
				spec.Name.Name,
				reembedded,
				reembedded,
			),
		}}
	}

	key := referenceKey(
		packageDirectory(file),
		spec.Name.Name,
	)
	isOnlyEmbedded := spec.Name.IsExported() && embedded[key] && !plain[key]
	if !isOnlyEmbedded {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"interface %s is only ever embedded: inline its methods into the interface that uses it",
			spec.Name.Name,
		),
	}}
}

func soleEmbedded(
	spec *ast.TypeSpec,
) (name string, isReembedding bool) {
	body, isInterface := spec.Type.(*ast.InterfaceType)
	isSingleEmbed := isInterface &&
		len(body.Methods.List) == 1 &&
		len(body.Methods.List[0].Names) == 0
	if !isSingleEmbed {
		return "",
			false
	}

	member := body.Methods.List[0].Type
	switch unindexed(member).(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return types.ExprString(member),
			true
	}

	return "",
		false
}

func interfaceSpecs(
	file source.File,
) (specs []*ast.TypeSpec) {
	return slices.DeleteFunc(
		source.Nodes[*ast.TypeSpec](file.Syntax),
		isNotInterface,
	)
}

func isNotInterface(
	spec *ast.TypeSpec,
) (excluded bool) {
	_, isInterface := spec.Type.(*ast.InterfaceType)
	return !isInterface || spec.Assign.IsValid()
}

func isExempt(
	file source.File,
) (exempt bool) {
	return source.IsTest(file.Path) || source.IsMock(file.Rel) || !source.IsInternalTree(packageDirectory(file))
}

func declaredPackageNames(
	files []source.File,
) (names map[string]string) {
	names = map[string]string{}
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isTestFile,
	) {
		names[packageDirectory(file)] = file.Syntax.Name.Name
	}

	return names
}

func isTestFile(
	file source.File,
) (isTest bool) {
	return source.IsTest(file.Path)
}

func unindexed(
	expression ast.Expr,
) (inner ast.Expr) {
	switch typed := expression.(type) {
	case *ast.IndexExpr:
		return typed.X
	case *ast.IndexListExpr:
		return typed.X
	}

	return expression
}

func referenceKey(
	directory string,
	name string,
) (key string) {
	return directory + pathSeparator + name
}

func packageDirectory(
	file source.File,
) (directory string) {
	return filepath.ToSlash(filepath.Dir(file.Rel))
}

func importedDirectory(
	importPath string,
	modulePath string,
) (directory string, isLocal bool) {
	if modulePath == "" {
		return "",
			false
	}

	if importPath == modulePath {
		return currentDirectory,
			true
	}

	return strings.CutPrefix(
		importPath,
		modulePath+pathSeparator,
	)
}
