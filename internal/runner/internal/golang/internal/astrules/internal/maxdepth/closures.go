package maxdepth

import (
	"go/ast"
	"maps"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

type closures struct {
	file    source.File
	indents map[int]int
	uses    map[*ast.FuncLit][]*ast.Ident
}

func closureShifts(
	file source.File,
	root ast.Node,
) (shifts map[*ast.FuncLit]int) {
	bound := closures{
		file: file,
		indents: lineIndents(
			file,
			root,
		),
		uses: literalUses(root),
	}
	shifts = map[*ast.FuncLit]int{}
	for literal := range bound.uses {
		shifts[literal] = 0
	}

	return bound.settled(
		shifts,
		len(shifts),
	)
}

func (c closures) settled(
	shifts map[*ast.FuncLit]int,
	passes int,
) (final map[*ast.FuncLit]int) {
	if passes == 0 {
		return shifts
	}

	next := c.relaxed(shifts)
	if maps.Equal(
		next,
		shifts,
	) {
		return shifts
	}

	return c.settled(
		next,
		passes-1,
	)
}

func (c closures) relaxed(
	shifts map[*ast.FuncLit]int,
) (next map[*ast.FuncLit]int) {
	next = map[*ast.FuncLit]int{}
	for literal := range shifts {
		next[literal] = c.shift(
			literal,
			shifts,
		)
	}

	return next
}

func (c closures) shift(
	literal *ast.FuncLit,
	shifts map[*ast.FuncLit]int,
) (levels int) {
	levels = closureShift(
		literal,
		shifts,
	)
	written := c.indents[c.file.Line(literal)]
	for _, use := range c.uses[literal] {
		levels = max(
			levels,
			c.indents[c.file.Line(use)]+closureShift(
				use,
				shifts,
			)-written,
		)
	}

	return levels
}

func closureShift(
	node ast.Node,
	shifts map[*ast.FuncLit]int,
) (levels int) {
	for literal, shift := range shifts {
		levels = max(
			levels,
			enclosedShift(
				literal,
				shift,
				node,
			),
		)
	}

	return levels
}

func enclosedShift(
	literal *ast.FuncLit,
	shift int,
	node ast.Node,
) (levels int) {
	if !encloses(
		literal.Body,
		node,
	) {
		return 0
	}

	return shift
}

func encloses(
	outer ast.Node,
	inner ast.Node,
) (enclosing bool) {
	return outer.Pos() <= inner.Pos() && inner.End() <= outer.End()
}

func lineIndents(
	file source.File,
	root ast.Node,
) (indents map[int]int) {
	indents = map[int]int{}
	for _, node := range source.Nodes[ast.Node](root) {
		line := file.Line(node)
		indents[line] = leftmost(
			indents,
			line,
			file.Column(node),
		)
	}

	return indents
}

func leftmost(
	indents map[int]int,
	line int,
	column int,
) (smallest int) {
	current, isSeen := indents[line]
	if !isSeen {
		return column
	}

	return min(
		current,
		column,
	)
}

func literalUses(
	root ast.Node,
) (uses map[*ast.FuncLit][]*ast.Ident) {
	bound := boundLiterals(root)
	uses = map[*ast.FuncLit][]*ast.Ident{}
	for _, ident := range source.Nodes[*ast.Ident](root) {
		recordUses(
			uses,
			bound[ident.Obj],
			ident,
		)
	}

	return uses
}

func recordUses(
	uses map[*ast.FuncLit][]*ast.Ident,
	literals []*ast.FuncLit,
	ident *ast.Ident,
) {
	for _, literal := range literals {
		recordUse(
			uses,
			literal,
			ident,
		)
	}
}

func recordUse(
	uses map[*ast.FuncLit][]*ast.Ident,
	literal *ast.FuncLit,
	ident *ast.Ident,
) {
	if encloses(
		literal.Body,
		ident,
	) {
		return
	}

	uses[literal] = append(
		uses[literal],
		ident,
	)
}

func boundLiterals(
	root ast.Node,
) (bound map[*ast.Object][]*ast.FuncLit) {
	bound = map[*ast.Object][]*ast.FuncLit{}
	for _, assign := range source.Nodes[*ast.AssignStmt](root) {
		bindPairs(
			bound,
			assign.Lhs,
			assign.Rhs,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](root) {
		bindPairs(
			bound,
			names(spec),
			spec.Values,
		)
	}

	return bound
}

func names(
	spec *ast.ValueSpec,
) (expressions []ast.Expr) {
	for _, name := range spec.Names {
		expressions = append(
			expressions,
			name,
		)
	}

	return expressions
}

func bindPairs(
	bound map[*ast.Object][]*ast.FuncLit,
	targets []ast.Expr,
	values []ast.Expr,
) {
	if len(targets) != len(values) {
		return
	}

	for index, value := range values {
		bindLiteral(
			bound,
			targets[index],
			value,
		)
	}
}

func bindLiteral(
	bound map[*ast.Object][]*ast.FuncLit,
	target ast.Expr,
	value ast.Expr,
) {
	ident, isIdent := target.(*ast.Ident)
	literal, isLiteral := value.(*ast.FuncLit)
	isBinding := isIdent && isLiteral && ident.Obj != nil
	if !isBinding {
		return
	}

	bound[ident.Obj] = append(
		bound[ident.Obj],
		literal,
	)
}
