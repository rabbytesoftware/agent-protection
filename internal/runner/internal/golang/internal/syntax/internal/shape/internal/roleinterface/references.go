package roleinterface

import (
	"go/ast"
	"path"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	importQuotes = "\"`"
	blankImport  = "_"
	dotImport    = "."
)

func fileReferences(
	file source.File,
	aliases map[string]string,
	packageNames map[string]string,
) (references []reference) {
	members := embeddedMembers(file)
	for _, selector := range source.Nodes[*ast.SelectorExpr](file.Syntax) {
		references = append(
			references,
			qualifiedReference(
				selector,
				aliases,
				members,
			)...,
		)
	}

	directory := packageDirectory(file)
	if file.Syntax.Name.Name != packageNames[directory] {
		return references
	}

	skipped := skippedNames(file)
	for _, identifier := range source.Nodes[*ast.Ident](file.Syntax) {
		references = append(
			references,
			bareReference(
				identifier,
				directory,
				members,
				skipped,
			)...,
		)
	}

	return references
}

func qualifiedReference(
	selector *ast.SelectorExpr,
	aliases map[string]string,
	members map[ast.Node]bool,
) (references []reference) {
	qualifier, isIdentifier := selector.X.(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	directory, isImported := aliases[qualifier.Name]
	if !isImported {
		return nil
	}

	return []reference{{
		key: referenceKey(
			directory,
			selector.Sel.Name,
		),
		embedded: members[selector],
	}}
}

func bareReference(
	identifier *ast.Ident,
	directory string,
	members map[ast.Node]bool,
	skipped map[*ast.Ident]bool,
) (references []reference) {
	if skipped[identifier] {
		return nil
	}

	return []reference{{
		key: referenceKey(
			directory,
			identifier.Name,
		),
		embedded: members[identifier],
	}}
}

func embeddedMembers(
	file source.File,
) (members map[ast.Node]bool) {
	members = map[ast.Node]bool{}
	for _, field := range interfaceFields(file) {
		members[unindexed(field.Type)] = len(field.Names) == 0
	}

	return members
}

func interfaceFields(
	file source.File,
) (fields []*ast.Field) {
	for _, body := range source.Nodes[*ast.InterfaceType](file.Syntax) {
		fields = append(
			fields,
			body.Methods.List...,
		)
	}

	return fields
}

func skippedNames(
	file source.File,
) (names map[*ast.Ident]bool) {
	names = map[*ast.Ident]bool{file.Syntax.Name: true}
	for _, identifier := range declaredIdentifiers(file) {
		names[identifier] = true
	}

	return names
}

func declaredIdentifiers(
	file source.File,
) (identifiers []*ast.Ident) {
	for _, spec := range source.Nodes[*ast.TypeSpec](file.Syntax) {
		identifiers = append(
			identifiers,
			spec.Name,
		)
	}

	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		identifiers = append(
			identifiers,
			function.Name,
		)
	}

	for _, field := range source.Nodes[*ast.Field](file.Syntax) {
		identifiers = append(
			identifiers,
			field.Names...,
		)
	}

	for _, selector := range source.Nodes[*ast.SelectorExpr](file.Syntax) {
		identifiers = append(
			identifiers,
			selector.Sel,
		)
	}

	return identifiers
}

func fileAliases(
	file source.File,
	modulePath string,
	packageNames map[string]string,
) (aliases map[string]string) {
	aliases = map[string]string{}
	for _, spec := range file.Syntax.Imports {
		recordAlias(
			aliases,
			spec,
			modulePath,
			packageNames,
		)
	}

	return aliases
}

func recordAlias(
	aliases map[string]string,
	spec *ast.ImportSpec,
	modulePath string,
	packageNames map[string]string,
) {
	importPath := importedPath(spec)
	directory, isLocal := importedDirectory(
		importPath,
		modulePath,
	)
	alias := importAlias(
		spec,
		importPath,
		packageNames[directory],
	)
	isNamed := isLocal && alias != blankImport && alias != dotImport
	if isNamed {
		aliases[alias] = directory
	}
}

func importAlias(
	spec *ast.ImportSpec,
	importPath string,
	declared string,
) (alias string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	if declared != "" {
		return declared
	}

	return path.Base(importPath)
}

func importedPath(
	spec *ast.ImportSpec,
) (importPath string) {
	return strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
}
