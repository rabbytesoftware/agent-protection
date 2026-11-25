package publicapi

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	vocabularyFormat  = "%s %s in vocabulary package %s: %s"
	behaviorRemedy    = "vocabulary packages hold declarations only; move behavior into the package that uses it"
	declarationRemedy = "vocabulary packages hold only exported types, constants and Err sentinels; move it into the package that uses it"
	renderingFormat   = "error method %s.%s in package %s, which an API file re-exports: a re-exported package holds declarations only, and Error, Unwrap, Is, As and String hold a single return of a field, a literal, a concatenation or one fmt.Sprintf, fmt.Errorf or errors call; move the logic into the package that uses it"
)

func vocabularyFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, flagged := range violations(declaration) {
		findings = append(
			findings,
			report.Finding{
				Rule:     vocabularyRule,
				Location: file.Location(flagged.node),
				Snippet: fmt.Sprintf(
					vocabularyFormat,
					flagged.kind,
					flagged.name,
					found.packageName,
					vocabularyRemedy(flagged),
				),
			},
		)
	}

	return findings
}

func vocabularyRemedy(
	flagged violation,
) (remedy string) {
	isBehavior := flagged.kind == funcKind || flagged.kind == methodKind || flagged.kind == errorMethodKind
	if isBehavior {
		return behaviorRemedy
	}

	return declarationRemedy
}

func vocabularyNames() (names []string) {
	return []string{
		"models",
		"model",
		"contracts",
		"contract",
		"types",
		"dto",
		"ports",
		"api",
		"options",
		"entities",
		"schema",
	}
}

func aliasedAcross(
	modulePath string,
	files []source.File,
	scopes []scope,
) (directories []string) {
	for index, file := range files {
		directories = append(
			directories,
			aliasedPackages(
				modulePath,
				file,
				scopes[index],
			)...,
		)
	}

	return directories
}

func aliasedPackages(
	modulePath string,
	file source.File,
	found scope,
) (directories []string) {
	if !found.api {
		return nil
	}

	for _, declaration := range file.Syntax.Decls {
		directories = append(
			directories,
			declarationAliases(
				modulePath,
				file.Syntax,
				declaration,
			)...,
		)
	}

	return directories
}

func declarationAliases(
	modulePath string,
	syntax *ast.File,
	declaration ast.Decl,
) (directories []string) {
	for _, spec := range slices.DeleteFunc(
		typeSpecs(declaration),
		isNotExportedAlias,
	) {
		directories = append(
			directories,
			aliasedDirectories(
				modulePath,
				syntax,
				aliasTarget(spec.Type),
			)...,
		)
	}

	return directories
}

func aliasedDirectories(
	modulePath string,
	syntax *ast.File,
	target ast.Expr,
) (directories []string) {
	selector, isSelector := target.(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}

	qualifier, isIdentifier := selector.X.(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	for _, imported := range qualifiedPaths(
		syntax,
		qualifier.Name,
	) {
		directories = append(
			directories,
			internalDirectory(
				modulePath,
				imported,
			)...,
		)
	}

	return directories
}

func internalDirectory(
	modulePath string,
	imported string,
) (directories []string) {
	directory, inModule := strings.CutPrefix(
		imported,
		modulePath+"/",
	)
	isInternal := inModule && source.IsInternalTree(directory)
	if !isInternal {
		return nil
	}

	return []string{directory}
}

func renderingFindings(
	files []source.File,
	aliased []string,
) (findings []report.Finding) {
	reExported := without(
		aliased,
		behavingDirectories(files),
	)
	for _, file := range files {
		findings = append(
			findings,
			fileRenderingFindings(
				file,
				reExported,
			)...,
		)
	}

	return findings
}

func behavingDirectories(
	files []source.File,
) (directories []string) {
	for _, file := range files {
		directories = append(
			directories,
			behavingDirectory(file)...,
		)
	}

	return directories
}

func behavingDirectory(
	file source.File,
) (directories []string) {
	if !slices.ContainsFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotRenderingMethod,
	) {
		return nil
	}

	return []string{filepath.ToSlash(filepath.Dir(file.Rel))}
}

func isNotRenderingMethod(
	function *ast.FuncDecl,
) (excluded bool) {
	return function.Recv == nil || !slices.Contains(
		source.ErrorMethods(),
		function.Name.Name,
	)
}

func fileRenderingFindings(
	file source.File,
	aliased []string,
) (findings []report.Finding) {
	if !slices.Contains(
		aliased,
		filepath.ToSlash(filepath.Dir(file.Rel)),
	) {
		return nil
	}

	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			renderingFinding(
				file,
				function,
			)...,
		)
	}

	return findings
}

func renderingFinding(
	file source.File,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	isLogic := function.Recv != nil && slices.Contains(
		source.ErrorMethods(),
		function.Name.Name,
	) && !source.IsPlainRendering(function.Body)
	if !isLogic {
		return nil
	}

	return []report.Finding{{
		Rule:     vocabularyRule,
		Location: file.Location(function),
		Snippet: fmt.Sprintf(
			renderingFormat,
			receiverType(function.Recv.List[0].Type),
			function.Name.Name,
			file.Syntax.Name.Name,
		),
	}}
}
