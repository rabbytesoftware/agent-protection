package deadvalue

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	deadRule    = "R69"
	proseRule   = "R07"
	blankName   = "_"
	iotaName    = "iota"
	localScope  = "<local>"
	deadFormat  = "%s %s is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message"
	proseFormat = "%s %s holds prose nothing reads (%q): a sentence parked in a string is a comment; delete it and make the code say it, or pass it to where a user sees it"
)

func Deadvalue(
	declared []source.File,
	referencing []source.File,
) (findings []report.Finding) {
	used := usages(referencing)
	for _, file := range slices.DeleteFunc(
		slices.Clone(declared),
		isExcluded,
	) {
		findings = append(
			findings,
			fileFindings(
				file,
				used,
			)...,
		)
	}

	return findings
}

func isExcluded(
	file source.File,
) (excluded bool) {
	return source.IsTest(file.Path) || source.IsMock(file.Rel) || file.IsGenerated
}

func fileFindings(
	file source.File,
	used usage,
) (findings []report.Finding) {
	for _, found := range slices.Concat(
		packageCandidates(
			file,
			used,
		),
		localCandidates(file),
	) {
		findings = append(
			findings,
			candidateFindings(
				file,
				found,
			)...,
		)
	}

	return findings
}

func candidateFindings(
	file source.File,
	found candidate,
) (findings []report.Finding) {
	isPublic := found.name.IsExported() &&
		found.directory != localScope &&
		!source.IsInternalTree(found.directory)
	if found.referenced ||
		isPublic {
		return nil
	}

	excerpt, isProse := proseOf(found.value)
	if isProse {
		return []report.Finding{{
			Rule:     proseRule,
			Location: file.Location(found.name),
			Snippet: fmt.Sprintf(
				proseFormat,
				found.kind,
				found.name.Name,
				excerpt,
			),
		}}
	}

	return []report.Finding{{
		Rule:     deadRule,
		Location: file.Location(found.name),
		Snippet: fmt.Sprintf(
			deadFormat,
			found.kind,
			found.name.Name,
		),
	}}
}

func packageCandidates(
	file source.File,
	used usage,
) (found []candidate) {
	directory := directoryOf(file)
	for _, declaration := range file.Syntax.Decls {
		found = append(
			found,
			groupCandidates(
				generalOf(declaration),
				used.names[directory],
				used.selected,
				directory,
			)...,
		)
	}

	return found
}

func localCandidates(
	file source.File,
) (found []candidate) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		found = append(
			found,
			functionCandidates(function)...,
		)
	}

	return found
}

func functionCandidates(
	function *ast.FuncDecl,
) (found []candidate) {
	if function.Body == nil {
		return nil
	}

	used := usage{
		names:    map[string]map[string]bool{},
		selected: map[string]bool{},
	}
	record(
		used,
		localScope,
		function.Body,
	)
	for _, statement := range source.Nodes[*ast.DeclStmt](function.Body) {
		found = append(
			found,
			groupCandidates(
				generalOf(statement.Decl),
				used.names[localScope],
				nil,
				localScope,
			)...,
		)
	}

	return found
}

func groupCandidates(
	general *ast.GenDecl,
	names map[string]bool,
	selected map[string]bool,
	directory string,
) (found []candidate) {
	isValueGroup := general != nil && (general.Tok == token.CONST || general.Tok == token.VAR) && !isEnumeration(general)
	if !isValueGroup {
		return nil
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](general) {
		found = append(
			found,
			specCandidates(
				general.Tok.String(),
				spec,
				names,
				selected,
				directory,
			)...,
		)
	}

	return found
}

func specCandidates(
	kind string,
	spec *ast.ValueSpec,
	names map[string]bool,
	selected map[string]bool,
	directory string,
) (found []candidate) {
	for index, name := range spec.Names {
		found = append(
			found,
			nameCandidates(
				candidate{
					kind: kind,
					name: name,
					value: valueAt(
						spec,
						index,
					),
					referenced: names[name.Name] || name.IsExported() && selected[name.Name],
					directory:  directory,
				},
			)...,
		)
	}

	return found
}

func nameCandidates(
	found candidate,
) (named []candidate) {
	if found.name.Name == blankName {
		return nil
	}

	return []candidate{found}
}

func generalOf(
	declaration ast.Decl,
) (general *ast.GenDecl) {
	general, isGeneral := declaration.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	return general
}

func valueAt(
	spec *ast.ValueSpec,
	index int,
) (value ast.Expr) {
	if len(spec.Values) != len(spec.Names) {
		return nil
	}

	return spec.Values[index]
}

func isEnumeration(
	general *ast.GenDecl,
) (enumeration bool) {
	for _, identifier := range source.Nodes[*ast.Ident](general) {
		enumeration = enumeration || identifier.Name == iotaName
	}

	return enumeration
}
