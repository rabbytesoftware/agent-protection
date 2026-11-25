package nostutter

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R23"

func Nostutter(
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
	for _, identifier := range declaredNames(file.Syntax) {
		findings = append(
			findings,
			identifierFindings(
				file,
				identifier,
			)...,
		)
	}

	return findings
}

func identifierFindings(
	file source.File,
	identifier *ast.Ident,
) (findings []report.Finding) {
	packageName := file.Syntax.Name.Name
	name := strings.ToLower(identifier.Name)
	prefix := strings.ToLower(packageName)
	isStutter := identifier.IsExported() && strings.HasPrefix(
		name,
		prefix,
	) && name != prefix
	if !isStutter {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(identifier),
		Snippet: fmt.Sprintf(
			"%s.%s stutters the package name: drop the repeated %q",
			packageName,
			identifier.Name,
			packageName,
		),
	}}
}

func declaredNames(
	syntax *ast.File,
) (identifiers []*ast.Ident) {
	for _, decl := range syntax.Decls {
		identifiers = append(
			identifiers,
			typeNames(decl)...,
		)
	}

	for _, function := range slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](syntax),
		isMethod,
	) {
		identifiers = append(
			identifiers,
			function.Name,
		)
	}

	return identifiers
}

func typeNames(
	decl ast.Decl,
) (identifiers []*ast.Ident) {
	general, isGeneral := decl.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	for _, spec := range source.Nodes[*ast.TypeSpec](general) {
		identifiers = append(
			identifiers,
			spec.Name,
		)
	}

	return identifiers
}

func isMethod(
	function *ast.FuncDecl,
) (method bool) {
	return function.Recv != nil
}
