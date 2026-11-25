package source

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

const (
	sentinelPrefix  = "Err"
	errorsQualifier = "errors"
	blankName       = "_"
)

func Nodes[T ast.Node](
	root ast.Node,
) (nodes []T) {
	for node := range ast.Preorder(root) {
		nodes = appendMatching(
			nodes,
			node,
		)
	}

	return nodes
}

func DeclarationPackages(
	files []File,
) (directories []string) {
	behaving := []string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		directories = append(
			directories,
			directory,
		)
		behaving = append(
			behaving,
			behavingDirectory(
				file,
				directory,
			)...,
		)
	}

	slices.Sort(directories)
	return withoutAny(
		slices.Compact(directories),
		behaving,
	)
}

func ErrorMethods() (names []string) {
	return []string{
		"Error",
		"Unwrap",
		"Is",
		"As",
		"String",
	}
}

func IsPlainRendering(
	body *ast.BlockStmt,
) (isPlain bool) {
	if body == nil {
		return true
	}

	if len(body.List) != 1 {
		return false
	}

	statement, isReturn := body.List[0].(*ast.ReturnStmt)
	return isReturn && len(statement.Results) <= 1 && !slices.ContainsFunc(
		statement.Results,
		isNotPlainResult,
	)
}

func isNotPlainResult(
	expression ast.Expr,
) (excluded bool) {
	call, isCall := ast.Unparen(expression).(*ast.CallExpr)
	isFormatting := isCall && isFormattingCall(call)
	if isFormatting {
		return false
	}

	return !isOperand(expression)
}

func isOperand(
	expression ast.Expr,
) (matches bool) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.BasicLit, *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return isOperand(typed.X)
	case *ast.BinaryExpr:
		return typed.Op == token.ADD && isOperand(typed.X) && isOperand(typed.Y)
	case *ast.CallExpr:
		return isConversion(typed) || isRenderingCall(typed)
	}

	return false
}

func isConversion(
	call *ast.CallExpr,
) (matches bool) {
	converted, isIdentifier := ast.Unparen(call.Fun).(*ast.Ident)
	return isIdentifier && len(call.Args) == 1 && slices.Contains(
		conversionTypes(),
		converted.Name,
	) && isOperand(call.Args[0])
}

func isRenderingCall(
	call *ast.CallExpr,
) (matches bool) {
	method, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	return isSelector && len(call.Args) == 0 && slices.Contains(
		renderingMethods(),
		method.Sel.Name,
	) && isOperand(method.X)
}

func isFormattingCall(
	call *ast.CallExpr,
) (matches bool) {
	function, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdentifier := function.X.(*ast.Ident)
	isFormatter := isIdentifier && (qualifier.Name == errorsQualifier || slices.Contains(
		formattingFunctions(),
		qualifier.Name+"."+function.Sel.Name,
	))
	return isFormatter && !slices.ContainsFunc(
		call.Args,
		isNotOperand,
	)
}

func isNotOperand(
	expression ast.Expr,
) (excluded bool) {
	return !isOperand(expression)
}

func conversionTypes() (names []string) {
	return []string{
		"string",
		"error",
		"bool",
		"int",
		"int64",
		"uint",
		"uint64",
		"float64",
		"rune",
		"byte",
	}
}

func renderingMethods() (names []string) {
	return []string{"Error", "String"}
}

func formattingFunctions() (names []string) {
	return []string{"fmt.Sprintf", "fmt.Errorf", "fmt.Sprint"}
}

func appendMatching[T ast.Node](
	nodes []T,
	node ast.Node,
) (matching []T) {
	typed, matches := node.(T)
	if !matches {
		return nodes
	}

	return append(
		nodes,
		typed,
	)
}

func behavingDirectory(
	file File,
	directory string,
) (behaving []string) {
	if !slices.ContainsFunc(
		file.Syntax.Decls,
		isBehavior,
	) {
		return nil
	}

	return []string{directory}
}

func isBehavior(
	declaration ast.Decl,
) (behaves bool) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return typed.Recv == nil || !slices.Contains(
			ErrorMethods(),
			typed.Name.Name,
		) || !IsPlainRendering(typed.Body)
	case *ast.GenDecl:
		return typed.Tok == token.VAR && slices.ContainsFunc(
			valueNames(typed),
			isStateName,
		)
	}

	return false
}

func isStateName(
	name *ast.Ident,
) (isState bool) {
	isSentinel := name.Name == "_" || name.IsExported() && strings.HasPrefix(
		name.Name,
		sentinelPrefix,
	)
	return !isSentinel
}

func valueNames(
	general *ast.GenDecl,
) (names []*ast.Ident) {
	for _, spec := range Nodes[*ast.ValueSpec](general) {
		names = append(
			names,
			spec.Names...,
		)
	}

	return names
}

func withoutAny(
	directories []string,
	excluded []string,
) (kept []string) {
	for _, directory := range directories {
		kept = append(
			kept,
			absentFrom(
				excluded,
				directory,
			)...,
		)
	}

	return kept
}

func absentFrom(
	excluded []string,
	directory string,
) (kept []string) {
	if slices.Contains(
		excluded,
		directory,
	) {
		return nil
	}

	return []string{directory}
}

func usesAny(
	identifiers []*ast.Ident,
	names []string,
) (uses bool) {
	for _, identifier := range identifiers {
		uses = uses || slices.Contains(
			names,
			identifier.Name,
		)
	}

	return uses
}

func declaredNames(
	syntax *ast.File,
) (names []string) {
	for _, declaration := range syntax.Decls {
		names = append(
			names,
			declarationNames(declaration)...,
		)
	}

	return slices.DeleteFunc(
		names,
		isBlankName,
	)
}

func declarationNames(
	declaration ast.Decl,
) (names []string) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return functionName(typed)
	case *ast.GenDecl:
		return specNames(typed)
	}

	return nil
}

func functionName(
	function *ast.FuncDecl,
) (names []string) {
	if function.Recv != nil {
		return nil
	}

	return []string{function.Name.Name}
}

func specNames(
	general *ast.GenDecl,
) (names []string) {
	for _, spec := range Nodes[*ast.TypeSpec](general) {
		names = append(
			names,
			spec.Name.Name,
		)
	}

	for _, name := range valueNames(general) {
		names = append(
			names,
			name.Name,
		)
	}

	return names
}

func foreignMethods(
	syntax *ast.File,
) (names []string) {
	declared := declaredNames(syntax)
	for _, function := range Nodes[*ast.FuncDecl](syntax) {
		names = append(
			names,
			foreignMethod(
				function,
				declared,
			)...,
		)
	}

	return names
}

func foreignMethod(
	function *ast.FuncDecl,
	declared []string,
) (names []string) {
	isForeign := function.Recv != nil && len(function.Recv.List) > 0 && !slices.Contains(
		declared,
		receiverName(function.Recv.List[0].Type),
	)
	if !isForeign {
		return nil
	}

	return []string{function.Name.Name}
}

func receiverName(
	expression ast.Expr,
) (name string) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverName(typed.X)
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	}

	return blankName
}

func selectedNames(
	syntax *ast.File,
) (names []*ast.Ident) {
	for _, selector := range Nodes[*ast.SelectorExpr](syntax) {
		names = append(
			names,
			selector.Sel,
		)
	}

	return names
}

func qualifiedNames(
	syntax *ast.File,
	qualifier string,
) (names []*ast.Ident) {
	for _, selector := range Nodes[*ast.SelectorExpr](syntax) {
		names = append(
			names,
			qualifiedName(
				selector,
				qualifier,
			)...,
		)
	}

	return names
}

func qualifiedName(
	selector *ast.SelectorExpr,
	qualifier string,
) (names []*ast.Ident) {
	identifier, isIdentifier := selector.X.(*ast.Ident)
	isQualified := isIdentifier && identifier.Obj == nil && identifier.Name == qualifier
	if !isQualified {
		return nil
	}

	return []*ast.Ident{selector.Sel}
}

func isBlankName(
	name string,
) (excluded bool) {
	return name == blankName
}
