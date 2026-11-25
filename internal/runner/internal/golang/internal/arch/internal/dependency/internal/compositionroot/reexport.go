package compositionroot

import (
	"go/ast"
	"go/token"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func isReExport(
	owner module,
	imported string,
	file source.File,
	spec *ast.ImportSpec,
) (matches bool) {
	if !slices.Contains(
		owner.vocabularies,
		filepath.ToSlash(imported),
	) {
		return false
	}

	qualifier := importName(spec)
	uses := len(qualifiedSelectors(
		file.Syntax,
		qualifier,
	))
	reExported := 0
	for _, node := range reExports(file.Syntax) {
		reExported += len(qualifiedSelectors(
			node,
			qualifier,
		))
	}

	return uses > 0 && uses == reExported
}

func importName(
	spec *ast.ImportSpec,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(strings.Trim(
		spec.Path.Value,
		importQuotes,
	))
}

func qualifiedSelectors(
	root ast.Node,
	qualifier string,
) (selectors []*ast.SelectorExpr) {
	for _, selector := range source.Nodes[*ast.SelectorExpr](root) {
		selectors = append(
			selectors,
			qualifiedBy(
				selector,
				qualifier,
			)...,
		)
	}

	return selectors
}

func qualifiedBy(
	selector *ast.SelectorExpr,
	qualifier string,
) (selectors []*ast.SelectorExpr) {
	identifier, isIdentifier := selector.X.(*ast.Ident)
	isQualified := isIdentifier && identifier.Name == qualifier
	if !isQualified {
		return nil
	}

	return []*ast.SelectorExpr{selector}
}

func reExports(
	syntax *ast.File,
) (nodes []ast.Node) {
	for _, general := range source.Nodes[*ast.GenDecl](syntax) {
		nodes = append(
			nodes,
			declarationReExports(general)...,
		)
	}

	return nodes
}

func declarationReExports(
	general *ast.GenDecl,
) (nodes []ast.Node) {
	switch general.Tok {
	case token.TYPE:
		return aliasTargets(general)
	case token.CONST:
		return []ast.Node{general}
	case token.VAR:
		return sentinelTargets(general)
	}

	return nil
}

func sentinelTargets(
	general *ast.GenDecl,
) (nodes []ast.Node) {
	for _, spec := range source.Nodes[*ast.ValueSpec](general) {
		nodes = append(
			nodes,
			sentinelTarget(spec)...,
		)
	}

	return nodes
}

func sentinelTarget(
	spec *ast.ValueSpec,
) (nodes []ast.Node) {
	isAliasing := len(spec.Values) == len(spec.Names) && !slices.ContainsFunc(
		spec.Names,
		isNotSentinel,
	)
	if !isAliasing {
		return nil
	}

	for _, value := range spec.Values {
		nodes = append(
			nodes,
			sentinelSelector(value)...,
		)
	}

	return nodes
}

func sentinelSelector(
	value ast.Expr,
) (nodes []ast.Node) {
	selector, isSelector := value.(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}

	_, isQualified := selector.X.(*ast.Ident)
	if !isQualified ||
		isNotSentinel(selector.Sel) {
		return nil
	}

	return []ast.Node{selector}
}

func isNotSentinel(
	name *ast.Ident,
) (excluded bool) {
	return !name.IsExported() || !strings.HasPrefix(
		name.Name,
		sentinelPrefix,
	)
}

func aliasTargets(
	general *ast.GenDecl,
) (nodes []ast.Node) {
	for _, spec := range source.Nodes[*ast.TypeSpec](general) {
		nodes = append(
			nodes,
			aliasTarget(spec)...,
		)
	}

	return nodes
}

func aliasTarget(
	spec *ast.TypeSpec,
) (nodes []ast.Node) {
	if !spec.Assign.IsValid() {
		return nil
	}

	return []ast.Node{spec.Type}
}
