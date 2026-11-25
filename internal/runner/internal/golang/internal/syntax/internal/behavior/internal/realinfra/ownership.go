package realinfra

import (
	"go/ast"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	newBuiltin   = "new"
	importQuotes = "\"`"
	blankImport  = "_"
	dotImport    = "."
)

func nodes[T ast.Node](
	root ast.Node,
) (found []T) {
	if root == nil {
		return nil
	}

	return source.Nodes[T](root)
}

func directoryScopes(
	tests []source.File,
) (scopes map[string]scope) {
	scopes = map[string]scope{}
	for _, file := range tests {
		directory := filepath.Dir(file.Rel)
		scopes[directory] = withFile(
			scopes[directory],
			file,
		)
	}

	return scopes
}

func withFile(
	known scope,
	file source.File,
) (grown scope) {
	local := scope{imports: importedPaths(file.Syntax)}
	grown = known
	grown.fields = map[string]string{}
	maps.Copy(
		grown.fields,
		known.fields,
	)
	maps.Copy(
		grown.fields,
		fieldKinds(
			local,
			file.Syntax,
		),
	)
	grown.makers = append(
		slices.Clone(known.makers),
		makers(
			local,
			file.Syntax,
		)...,
	)
	return grown
}

func fieldKinds(
	local scope,
	syntax *ast.File,
) (kinds map[string]string) {
	kinds = map[string]string{}
	for _, field := range nodes[*ast.Field](syntax) {
		maps.Copy(
			kinds,
			namedKinds(
				field.Names,
				local.typeKind(field.Type),
			),
		)
	}

	return kinds
}

func namedKinds(
	names []*ast.Ident,
	kind string,
) (kinds map[string]string) {
	kinds = map[string]string{}
	if kind == "" {
		return kinds
	}

	for _, name := range names {
		kinds[name.Name] = kind
	}

	return kinds
}

func makers(
	local scope,
	syntax *ast.File,
) (names []string) {
	for _, function := range nodes[*ast.FuncDecl](syntax) {
		names = append(
			names,
			makerName(
				local,
				function,
			)...,
		)
	}

	return names
}

func makerName(
	local scope,
	function *ast.FuncDecl,
) (names []string) {
	isMaker := function.Recv == nil && function.Body != nil && (local.returnsServer(function.Type) || slices.ContainsFunc(
		nodes[*ast.CallExpr](function.Body),
		local.isOwnedSeed,
	))
	if !isMaker {
		return nil
	}

	return []string{function.Name.Name}
}

func functionScope(
	base scope,
	imports map[string]string,
	function *ast.FuncDecl,
) (found scope) {
	found = base
	found.imports = imports
	found.owned = nil
	found.kinds = map[string]string{}
	bindings := parameterBindings(function.Type)
	for _, literal := range nodes[*ast.FuncLit](function.Body) {
		bindings = append(
			bindings,
			parameterBindings(literal.Type)...,
		)
	}

	bindings = append(
		bindings,
		bodyBindings(function.Body)...,
	)
	for grown := true; grown; {
		found, grown = found.bound(bindings)
	}

	return found
}

func parameterBindings(
	function *ast.FuncType,
) (bindings []binding) {
	for _, field := range function.Params.List {
		bindings = append(
			bindings,
			binding{
				names:    identifierNames(field.Names),
				declared: field.Type,
			},
		)
	}

	return bindings
}

func bodyBindings(
	body *ast.BlockStmt,
) (bindings []binding) {
	for _, assignment := range nodes[*ast.AssignStmt](body) {
		bindings = append(
			bindings,
			assignmentBindings(assignment)...,
		)
	}

	for _, spec := range nodes[*ast.ValueSpec](body) {
		bindings = append(
			bindings,
			specBindings(spec)...,
		)
	}

	return bindings
}

func assignmentBindings(
	assignment *ast.AssignStmt,
) (bindings []binding) {
	if len(assignment.Rhs) == 1 {
		return []binding{{
			names: expressionNames(assignment.Lhs),
			value: assignment.Rhs[0],
		}}
	}

	for index, value := range assignment.Rhs {
		bindings = append(
			bindings,
			binding{
				names: expressionNames(assignment.Lhs[index : index+1]),
				value: value,
			},
		)
	}

	return bindings
}

func specBindings(
	spec *ast.ValueSpec,
) (bindings []binding) {
	names := identifierNames(spec.Names)
	if len(spec.Values) != len(names) {
		return []binding{{
			names:    names,
			value:    firstValue(spec.Values),
			declared: spec.Type,
		}}
	}

	for index, value := range spec.Values {
		bindings = append(
			bindings,
			binding{
				names:    names[index : index+1],
				value:    value,
				declared: spec.Type,
			},
		)
	}

	return bindings
}

func firstValue(
	values []ast.Expr,
) (value ast.Expr) {
	if len(values) == 0 {
		return nil
	}

	return values[0]
}

func expressionNames(
	expressions []ast.Expr,
) (names []string) {
	for _, expression := range expressions {
		names = append(
			names,
			expressionName(expression)...,
		)
	}

	return names
}

func expressionName(
	expression ast.Expr,
) (names []string) {
	identifier, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return nil
	}

	return []string{identifier.Name}
}

func identifierNames(
	identifiers []*ast.Ident,
) (names []string) {
	for _, identifier := range identifiers {
		names = append(
			names,
			identifier.Name,
		)
	}

	return names
}

func selectedNames(
	expression ast.Expr,
) (selected []*ast.Ident) {
	for _, selector := range nodes[*ast.SelectorExpr](expression) {
		selected = append(
			selected,
			selector.Sel,
		)
	}

	return selected
}

func isLocalAddress(
	selector *ast.SelectorExpr,
) (local bool) {
	receiver, isIdent := selector.X.(*ast.Ident)
	return isIdent && receiver.Obj != nil && selector.Sel.Name == addressMethod
}

func isEphemeralAddress(
	argument ast.Expr,
) (ephemeral bool) {
	literal, isLiteral := argument.(*ast.BasicLit)
	return isLiteral && literal.Kind == token.STRING && strings.HasSuffix(
		strings.Trim(
			literal.Value,
			stringQuotes,
		),
		ephemeralPort,
	)
}

func hasAnyPrefix(
	name string,
	prefixes []string,
) (prefixed bool) {
	for _, prefix := range prefixes {
		prefixed = prefixed || strings.HasPrefix(
			name,
			prefix,
		)
	}

	return prefixed
}

func firstKind(
	declared string,
	assigned string,
) (kind string) {
	if declared != "" {
		return declared
	}

	return assigned
}

func appendMissing(
	names []string,
	name string,
) (extended []string) {
	if name == blankImport ||
		slices.Contains(
			names,
			name,
		) {
		return names
	}

	return append(
		names,
		name,
	)
}

func importedPaths(
	syntax *ast.File,
) (imports map[string]string) {
	imports = map[string]string{}
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(
			spec.Path.Value,
			importQuotes,
		)
		imports[importName(
			spec,
			importPath,
		)] = importPath
	}

	delete(
		imports,
		blankImport,
	)
	delete(
		imports,
		dotImport,
	)
	return imports
}

func importName(
	spec *ast.ImportSpec,
	importPath string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	elements := strings.Split(
		normalizedPath(importPath),
		pathSeparator,
	)
	name = strings.TrimPrefix(
		elements[len(elements)-1],
		goPrefix,
	)
	for _, suffix := range goSuffixes() {
		name = strings.TrimSuffix(
			name,
			suffix,
		)
	}

	return name
}

func goSuffixes() (suffixes []string) {
	return []string{
		"-go",
		".go",
	}
}
