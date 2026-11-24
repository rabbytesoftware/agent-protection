package paramperline

import (
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Edits(
	file source.File,
) (edits []source.Edit) {
	for _, node := range source.Nodes[ast.Node](file.Syntax) {
		edits = append(
			edits,
			nodeEdits(
				file,
				node,
			)...,
		)
	}

	return edits
}

func nodeEdits(
	file source.File,
	node ast.Node,
) (edits []source.Edit) {
	_, funcType := signature(node)
	if funcType == nil {
		return nil
	}

	return source.Breaks(
		file,
		funcType.Params.Opening,
		parameterExtents(funcType.Params),
		funcType.Params.Closing,
	)
}

func parameterExtents(
	params *ast.FieldList,
) (extents []source.Extent) {
	for _, field := range params.List {
		extents = append(
			extents,
			fieldExtents(field)...,
		)
	}

	return extents
}

func fieldExtents(
	field *ast.Field,
) (extents []source.Extent) {
	if len(field.Names) == 0 {
		return []source.Extent{{Start: field.Pos(), End: field.End()}}
	}

	for _, name := range field.Names {
		extents = append(
			extents,
			source.Extent{Start: name.Pos(), End: name.End()},
		)
	}

	extents[len(extents)-1].End = field.End()
	return extents
}
