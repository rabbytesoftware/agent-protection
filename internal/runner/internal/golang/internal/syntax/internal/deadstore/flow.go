package deadstore

import (
	"go/ast"
	"go/token"
	"strings"
)

const (
	errorName   = "err"
	errorSuffix = "Err"
)

type flow struct {
	scopes   []map[string]token.Pos
	depths   map[token.Pos]int
	current  reaching
	frames   []*frame
	loops    []ast.Node
	stores   map[token.Pos]store
	used     map[token.Pos]bool
	exited   map[token.Pos]bool
	killed   map[token.Pos]bool
	cycled   map[token.Pos]bool
	reads    map[token.Pos][]token.Pos
	excluded map[string]bool
	results  []token.Pos
}

func flowFor(
	excluded map[string]bool,
) (tracker *flow) {
	return &flow{
		depths:   map[token.Pos]int{},
		current:  reaching{},
		stores:   map[token.Pos]store{},
		used:     map[token.Pos]bool{},
		exited:   map[token.Pos]bool{},
		killed:   map[token.Pos]bool{},
		cycled:   map[token.Pos]bool{},
		reads:    map[token.Pos][]token.Pos{},
		excluded: excluded,
	}
}

func (f *flow) pushScope() {
	f.scopes = append(
		f.scopes,
		map[string]token.Pos{},
	)
}

func (f *flow) popScope() {
	for _, variable := range f.scopes[len(f.scopes)-1] {
		f.leave(variable)
	}

	f.scopes = f.scopes[:len(f.scopes)-1]
}

func (f *flow) leave(
	variable token.Pos,
) {
	for _, position := range f.current[variable] {
		f.exited[position] = true
	}

	delete(
		f.current,
		variable,
	)
}

func (f *flow) declare(
	name *ast.Ident,
) {
	isTracked := name != nil && isErrorName(name.Name) && !f.excluded[name.Name]
	if !isTracked {
		return
	}

	f.scopes[len(f.scopes)-1][name.Name] = name.Pos()
	f.depths[name.Pos()] = len(f.loops)
}

func (f *flow) declaredHere(
	name string,
) (declared bool) {
	_, declared = f.scopes[len(f.scopes)-1][name]
	return declared
}

func (f *flow) lookup(
	name string,
) (variable token.Pos) {
	for index := len(f.scopes) - 1; index >= 0 && variable == token.NoPos; index-- {
		variable = f.scopes[index][name]
	}

	return variable
}

func (f *flow) read(
	name *ast.Ident,
) {
	f.readVariable(
		f.lookup(name.Name),
		name.Pos(),
	)
}

func (f *flow) readVariable(
	variable token.Pos,
	position token.Pos,
) {
	if variable == token.NoPos {
		return
	}

	f.reads[variable] = append(
		f.reads[variable],
		position,
	)
	for _, stored := range f.current[variable] {
		f.used[stored] = true
	}
}

func (f *flow) assign(
	target *ast.Ident,
	value ast.Expr,
) {
	variable := f.lookup(target.Name)
	if variable == token.NoPos ||
		f.current == nil {
		return
	}

	for _, previous := range f.current[variable] {
		f.killed[previous] = true
	}

	f.stores[target.Pos()] = store{variable: variable, target: target, call: callName(value), loop: f.enclosingLoop(variable)}
	f.current[variable] = []token.Pos{target.Pos()}
}

func (f *flow) enclosingLoop(
	variable token.Pos,
) (loop ast.Node) {
	if len(f.loops) <= f.depths[variable] {
		return nil
	}

	return f.loops[len(f.loops)-1]
}

func (f *flow) exit() {
	for variable := range f.current {
		f.leave(variable)
	}

	f.current = nil
}

func (f *flow) breakOut() {
	innermost := f.frames[len(f.frames)-1]
	innermost.breaks = append(
		innermost.breaks,
		f.current,
	)
	f.current = nil
}

func (f *flow) continueLoop() {
	for index := len(f.frames) - 1; index >= 0; index-- {
		f.continueAt(f.frames[index])
	}

	f.current = nil
}

func (f *flow) continueAt(
	target *frame,
) {
	if target.loop == nil ||
		f.current == nil {
		return
	}

	target.continues = append(
		target.continues,
		f.current,
	)
	f.current = nil
}

func (f *flow) markCycled(
	loop ast.Node,
) {
	for _, stored := range f.current {
		f.markLoopStores(
			stored,
			loop,
		)
	}
}

func (f *flow) markLoopStores(
	stored []token.Pos,
	loop ast.Node,
) {
	for _, position := range stored {
		f.cycled[position] = f.cycled[position] || f.stores[position].loop == loop
	}
}

func (f *flow) readWithin(
	variable token.Pos,
	loop ast.Node,
) (read bool) {
	for _, position := range f.reads[variable] {
		read = read || loop.Pos() <= position && position < loop.End()
	}

	return read
}

func isErrorName(
	name string,
) (matches bool) {
	return name == errorName || strings.HasSuffix(
		name,
		errorSuffix,
	)
}
