package interfacesize

import (
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	structureFormat = "struct %s exposes %d methods through embedded %s%s (max %d): a struct that embeds interfaces is one large interface; return each small interface from its own constructor and let every consumer depend on the one it calls"
	ownFormat       = " plus %d own methods"
	functionFormat  = "struct %s exposes %d methods through %d exported func fields%s%s (max %d): a struct of exported func fields is an interface assembled by hand; declare an interface with the methods each consumer calls, implement it with methods on an unexported type and return it from its constructor"
	embeddedFormat  = " and embedded %s"
)

func structureFindings(
	file source.File,
	key string,
	indexed catalog,
) (findings []report.Finding) {
	declared, isStructure := indexed.structures[key]
	if !isStructure {
		return nil
	}

	members, largest := embeddedMembers(
		declared,
		indexed,
		[]string{key},
	)
	exposed, isBacked := structureSurface(
		indexed,
		key,
		nil,
	)
	isBloated := isBacked &&
		len(exposed) > maxMethods &&
		len(exposed) > largest
	if !isBloated {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(declared.owner.spec),
		Snippet: surfaceSnippet(
			declared,
			len(exposed),
			members,
			len(indexed.methods[key]),
			len(functionFields(
				declared,
				indexed,
			)),
		),
	}}
}

func surfaceSnippet(
	declared structure,
	exposed int,
	members []*ast.Field,
	own int,
	functions int,
) (snippet string) {
	if functions == 0 {
		return fmt.Sprintf(
			structureFormat,
			declared.owner.spec.Name.Name,
			exposed,
			embedNames(members),
			ownNote(own),
			maxMethods,
		)
	}

	return fmt.Sprintf(
		functionFormat,
		declared.owner.spec.Name.Name,
		exposed,
		functions,
		embeddedNote(members),
		ownNote(own),
		maxMethods,
	)
}

func embeddedNote(
	members []*ast.Field,
) (note string) {
	if len(members) == 0 {
		return ""
	}

	return fmt.Sprintf(
		embeddedFormat,
		embedNames(members),
	)
}

func functionFields(
	declared structure,
	indexed catalog,
) (names []string) {
	for _, field := range declared.body.Fields.List {
		names = append(
			names,
			exportedFunctions(
				declared.owner,
				field,
				indexed,
			)...,
		)
	}

	return names
}

func exportedFunctions(
	owner declaration,
	field *ast.Field,
	indexed catalog,
) (names []string) {
	if !isFunctionType(
		owner,
		field.Type,
		indexed,
	) {
		return nil
	}

	for _, name := range field.Names {
		names = appendExported(
			names,
			name,
		)
	}

	return names
}

func appendExported(
	names []string,
	name *ast.Ident,
) (extended []string) {
	if !name.IsExported() {
		return names
	}

	return append(
		names,
		name.Name,
	)
}

func isFunctionType(
	owner declaration,
	expression ast.Expr,
	indexed catalog,
) (isFunction bool) {
	_, isLiteral := ast.Unparen(expression).(*ast.FuncType)
	return isLiteral || indexed.functions[memberKey(
		owner,
		unindexed(ast.Unparen(expression)),
	)]
}

func ownNote(
	count int,
) (note string) {
	if count == 0 {
		return ""
	}

	return fmt.Sprintf(
		ownFormat,
		count,
	)
}

func embeddedMembers(
	declared structure,
	indexed catalog,
	visiting []string,
) (members []*ast.Field, largest int) {
	for _, field := range declared.body.Fields.List {
		promoted, _ := memberSurface(
			declared.owner,
			field,
			indexed,
			visiting,
		)
		members = append(
			members,
			contributing(
				declared.owner,
				field,
				indexed,
				visiting,
			)...,
		)
		largest = max(
			largest,
			len(promoted),
		)
	}

	return members,
		largest
}

func contributing(
	owner declaration,
	field *ast.Field,
	indexed catalog,
	visiting []string,
) (fields []*ast.Field) {
	promoted, isBacked := memberSurface(
		owner,
		field,
		indexed,
		visiting,
	)
	isSilent := len(promoted) == 0 &&
		!isBacked
	if isSilent {
		return nil
	}

	return []*ast.Field{field}
}

func structureSurface(
	indexed catalog,
	key string,
	visiting []string,
) (names []string, isBacked bool) {
	path := slices.Concat(
		visiting,
		[]string{key},
	)
	declared := indexed.structures[key]
	names = slices.Concat(
		indexed.methods[key],
		functionFields(
			declared,
			indexed,
		),
	)
	isBacked = len(names) > len(indexed.methods[key])
	for _, field := range declared.body.Fields.List {
		promoted, isInterface := memberSurface(
			declared.owner,
			field,
			indexed,
			path,
		)
		names = append(
			names,
			promoted...,
		)
		isBacked = isBacked || isInterface
	}

	slices.Sort(names)
	return slices.Compact(names),
		isBacked
}

func memberSurface(
	owner declaration,
	field *ast.Field,
	indexed catalog,
	visiting []string,
) (names []string, isBacked bool) {
	if len(field.Names) > 0 {
		return nil,
			false
	}

	embedded, isInterface := indexed.interfaces[memberKey(
		owner,
		unindexed(field.Type),
	)]
	if isInterface {
		return methodNames(
				embedded,
				indexed.interfaces,
				nil,
			),
			true
	}

	return embeddedStructSurface(
		owner,
		field,
		indexed,
		visiting,
	)
}

func embeddedStructSurface(
	owner declaration,
	field *ast.Field,
	indexed catalog,
	visiting []string,
) (names []string, isBacked bool) {
	key := memberKey(
		owner,
		unindexed(unpointered(field.Type)),
	)
	_, isStructure := indexed.structures[key]
	isExpandable := isStructure && !slices.Contains(
		visiting,
		key,
	)
	if !isExpandable {
		return nil,
			false
	}

	return structureSurface(
		indexed,
		key,
		visiting,
	)
}
