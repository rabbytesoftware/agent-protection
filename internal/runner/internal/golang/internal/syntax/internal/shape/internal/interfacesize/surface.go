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
		Snippet: fmt.Sprintf(
			structureFormat,
			declared.owner.spec.Name.Name,
			len(exposed),
			embedNames(members),
			ownNote(len(indexed.methods[key])),
			maxMethods,
		),
	}}
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
	names = slices.Clone(indexed.methods[key])
	declared := indexed.structures[key]
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
