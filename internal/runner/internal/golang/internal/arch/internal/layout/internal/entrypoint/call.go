package entrypoint

import (
	"go/ast"
	"path"
	"slices"
	"strings"
)

const (
	importQuotes  = "\"`"
	exitCall      = "os.Exit"
	printCall     = "fmt.Fprintln"
	stderrName    = "Stderr"
	osQualifier   = "os"
	pathSeparator = "/"
	domainDot     = "."
)

func callVerb(
	context delegation,
	call *ast.CallExpr,
) (verb string) {
	switch callee := calleeOf(call.Fun).(type) {
	case *ast.Ident:
		return identifierCallVerb(
			context,
			callee,
		)
	case *ast.SelectorExpr:
		return selectorCallVerb(
			context,
			call,
			callee,
		)
	case *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType:
		return "converts"
	}

	return noLogic
}

func calleeOf(
	function ast.Expr,
) (callee ast.Expr) {
	switch typed := ast.Unparen(function).(type) {
	case *ast.IndexExpr:
		return ast.Unparen(typed.X)
	case *ast.IndexListExpr:
		return ast.Unparen(typed.X)
	}

	return ast.Unparen(function)
}

func identifierCallVerb(
	context delegation,
	callee *ast.Ident,
) (verb string) {
	isDeclared := slices.Contains(
		context.unit.names,
		callee.Name,
	) || slices.Contains(
		context.locals,
		callee.Name,
	)
	if isDeclared {
		return noLogic
	}

	return "calls " + callee.Name
}

func selectorCallVerb(
	context delegation,
	call *ast.CallExpr,
	callee *ast.SelectorExpr,
) (verb string) {
	imported, isImport := qualifierPath(
		context,
		callee.X,
	)
	if !isImport {
		return noLogic
	}

	qualified := path.Base(imported) + "." + callee.Sel.Name
	isPermitted := !isStandard(
		context.unit.modulePath,
		imported,
	) || isAllowedCall(
		context.unit.role,
		imported+"."+callee.Sel.Name,
		call,
	)
	if isPermitted {
		return noLogic
	}

	return "calls " + qualified
}

func qualifierPath(
	context delegation,
	expression ast.Expr,
) (imported string, isImport bool) {
	qualifier, isIdentifier := expression.(*ast.Ident)
	isQualifier := isIdentifier && !slices.Contains(
		context.locals,
		qualifier.Name,
	)
	if !isQualifier {
		return "",
			false
	}

	paths := importsNamed(
		context.unit.file.Syntax,
		qualifier.Name,
	)
	if len(paths) == 0 {
		return "",
			false
	}

	return paths[0],
		true
}

func isAllowedCall(
	role string,
	qualified string,
	call *ast.CallExpr,
) (allowed bool) {
	if slices.Contains(
		allowedCalls(),
		qualified,
	) {
		return true
	}

	if role != mainRole {
		return false
	}

	return qualified == exitCall || qualified == printCall && isStderr(call.Args)
}

func isStderr(
	arguments []ast.Expr,
) (matches bool) {
	if len(arguments) == 0 {
		return false
	}

	selector, isSelector := ast.Unparen(arguments[0]).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdentifier := selector.X.(*ast.Ident)
	return isIdentifier && qualifier.Name == osQualifier && selector.Sel.Name == stderrName
}

func isStandard(
	modulePath string,
	imported string,
) (standard bool) {
	isOwnModule := imported == modulePath || strings.HasPrefix(
		imported,
		modulePath+pathSeparator,
	)
	first, _, _ := strings.Cut(
		imported,
		pathSeparator,
	)
	return !isOwnModule && !strings.Contains(
		first,
		domainDot,
	)
}

func importsNamed(
	syntax *ast.File,
	qualifier string,
) (paths []string) {
	for _, spec := range syntax.Imports {
		paths = append(
			paths,
			importNamed(
				spec,
				qualifier,
			)...,
		)
	}

	return paths
}

func importNamed(
	spec *ast.ImportSpec,
	qualifier string,
) (paths []string) {
	imported := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	name := path.Base(imported)
	if spec.Name != nil {
		name = spec.Name.Name
	}

	if name != qualifier {
		return nil
	}

	return []string{imported}
}

func allowedCalls() (qualified []string) {
	return []string{"errors.New", "errors.Join", "fmt.Errorf", "slices.Concat"}
}
