package assertion

import (
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

type catalog struct {
	declarations []declaration
	helpers      []string
	fields       map[string][]string
	values       map[string][]string
	directories  []string
}

func catalogOf(
	files []source.File,
) (known catalog) {
	known.helpers = failingInterfaces(files)
	known.fields = testingFields(
		files,
		known.helpers,
	)
	known.values = valueNames(files)
	for _, file := range files {
		known.declarations = append(
			known.declarations,
			fileDeclarations(file)...,
		)
		known.directories = append(
			known.directories,
			filepath.Dir(file.Rel),
		)
	}

	slices.Sort(known.directories)
	known.directories = slices.Compact(known.directories)
	return known
}

func (c catalog) indexOf(
	body *ast.BlockStmt,
) (index int) {
	bodies := []*ast.BlockStmt{}
	for _, known := range c.declarations {
		bodies = append(
			bodies,
			known.body,
		)
	}

	return slices.Index(
		bodies,
		body,
	)
}

func (c catalog) resolve(
	owner declaration,
	expression ast.Expr,
) (targets []int) {
	switch typed := unwrapped(expression).(type) {
	case *ast.Ident:
		return c.matching(
			[]string{filepath.Dir(owner.file.Rel)},
			typed.Name,
			isFunction,
		)
	case *ast.SelectorExpr:
		return c.resolveSelector(
			owner,
			typed,
		)
	}

	return nil
}

func (c catalog) resolveSelector(
	owner declaration,
	selector *ast.SelectorExpr,
) (targets []int) {
	qualifier, isIdent := selector.X.(*ast.Ident)
	importPath, isImported := owner.imports[identifierName(qualifier)]
	if isIdent &&
		isImported {
		return c.matching(
			[]string{c.directoryOf(importPath)},
			selector.Sel.Name,
			isFunction,
		)
	}

	return c.matching(
		c.visibleDirectories(owner),
		selector.Sel.Name,
		isMethod,
	)
}

func (c catalog) visibleDirectories(
	owner declaration,
) (directories []string) {
	directories = []string{filepath.Dir(owner.file.Rel)}
	for _, importPath := range owner.imports {
		directories = append(
			directories,
			c.directoryOf(importPath),
		)
	}

	return directories
}

func (c catalog) directoryOf(
	importPath string,
) (directory string) {
	for _, candidate := range c.directories {
		directory = longerMatch(
			directory,
			candidate,
			importPath,
		)
	}

	return directory
}

func (c catalog) matching(
	directories []string,
	name string,
	kind string,
) (targets []int) {
	for index, candidate := range c.declarations {
		targets = append(
			targets,
			matchingIndex(
				index,
				candidate,
				directories,
				name,
				kind,
			)...,
		)
	}

	return targets
}

func matchingIndex(
	index int,
	candidate declaration,
	directories []string,
	name string,
	kind string,
) (targets []int) {
	isMatch := candidate.name == name && kindOf(candidate) == kind && slices.Contains(
		directories,
		filepath.Dir(candidate.file.Rel),
	)
	if !isMatch {
		return nil
	}

	return []int{index}
}

func longerMatch(
	current string,
	candidate string,
	importPath string,
) (directory string) {
	isLonger := len(candidate) > len(current) && candidate != rootDirectory && strings.HasSuffix(
		pathSeparator+importPath,
		pathSeparator+filepath.ToSlash(candidate),
	)
	if !isLonger {
		return current
	}

	return candidate
}

func (c catalog) dynamicTargets(
	owner declaration,
) (targets []int) {
	for index, candidate := range c.declarations {
		targets = append(
			targets,
			c.dynamicIndex(
				owner,
				index,
				candidate,
			)...,
		)
	}

	return targets
}

func (c catalog) dynamicIndex(
	owner declaration,
	index int,
	candidate declaration,
) (targets []int) {
	isTarget := filepath.Dir(candidate.file.Rel) == filepath.Dir(owner.file.Rel) && !isTestName(candidate.name) && (c.isValuedHelper(candidate) || c.holdsTestingLiteral(candidate))
	if !isTarget {
		return nil
	}

	return []int{index}
}

func (c catalog) isValuedHelper(
	candidate declaration,
) (valued bool) {
	return takesTesting(
		c,
		candidate,
	) && slices.Contains(
		c.values[filepath.Dir(candidate.file.Rel)],
		candidate.name,
	)
}

func (c catalog) holdsTestingLiteral(
	candidate declaration,
) (holds bool) {
	alias := testingAlias(candidate.imports)
	for _, literal := range source.Nodes[*ast.FuncLit](candidate.body) {
		holds = holds || len(fieldNames(
			literal.Type.Params,
			alias,
			c.helpers,
		)) > 0
	}

	return holds
}
