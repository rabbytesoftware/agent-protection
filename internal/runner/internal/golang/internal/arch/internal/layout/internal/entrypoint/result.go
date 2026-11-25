package entrypoint

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const resultFormat = "container.go func %s returns %s: entry functions return an interface or type declared in package %s, a basic type or error, never a type from an internal package or a pointer into another package; declare the interface callers need in %s (or alias a declaration-only internal type there) and return the implementation through it"

func resultFindings(
	found unit,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Type.Results == nil {
		return nil
	}

	typeParameters := fieldNames(function.Type.TypeParams)
	for _, field := range function.Type.Results.List {
		findings = append(
			findings,
			fieldFindings(
				found,
				function,
				typeParameters,
				field,
			)...,
		)
	}

	return findings
}

func fieldFindings(
	found unit,
	function *ast.FuncDecl,
	typeParameters []string,
	field *ast.Field,
) (findings []report.Finding) {
	if len(exposures(
		found,
		typeParameters,
		field.Type,
	)) == 0 {
		return nil
	}

	return []report.Finding{{
		Rule:     resultRule,
		Location: found.file.Location(field.Type),
		Snippet: fmt.Sprintf(
			resultFormat,
			function.Name.Name,
			types.ExprString(field.Type),
			found.file.Syntax.Name.Name,
			found.directory,
		),
	}}
}

func exposures(
	found unit,
	typeParameters []string,
	expression ast.Expr,
) (exposed []ast.Expr) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.StarExpr:
		return pointerExposures(
			found,
			typeParameters,
			typed,
		)
	case *ast.ArrayType:
		return exposures(
			found,
			typeParameters,
			typed.Elt,
		)
	case *ast.MapType:
		return everyExposure(
			found,
			typeParameters,
			[]ast.Expr{typed.Key, typed.Value},
		)
	case *ast.ChanType:
		return exposures(
			found,
			typeParameters,
			typed.Value,
		)
	}

	return namedExposures(
		found,
		typeParameters,
		expression,
	)
}

func namedExposures(
	found unit,
	typeParameters []string,
	expression ast.Expr,
) (exposed []ast.Expr) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.IndexExpr:
		return everyExposure(
			found,
			typeParameters,
			[]ast.Expr{typed.X, typed.Index},
		)
	case *ast.IndexListExpr:
		return everyExposure(
			found,
			typeParameters,
			slices.Concat(
				[]ast.Expr{typed.X},
				typed.Indices,
			),
		)
	case *ast.Ident:
		return identifierExposures(
			found,
			typeParameters,
			typed,
		)
	case *ast.SelectorExpr:
		return selectorExposures(
			found,
			typed,
		)
	case *ast.InterfaceType:
		return nil
	}

	return []ast.Expr{expression}
}

func everyExposure(
	found unit,
	typeParameters []string,
	expressions []ast.Expr,
) (exposed []ast.Expr) {
	for _, expression := range expressions {
		exposed = append(
			exposed,
			exposures(
				found,
				typeParameters,
				expression,
			)...,
		)
	}

	return exposed
}

func pointerExposures(
	found unit,
	typeParameters []string,
	pointer *ast.StarExpr,
) (exposed []ast.Expr) {
	_, isForeign := calleeOf(pointer.X).(*ast.SelectorExpr)
	if isForeign {
		return []ast.Expr{pointer}
	}

	return exposures(
		found,
		typeParameters,
		pointer.X,
	)
}

func identifierExposures(
	found unit,
	typeParameters []string,
	identifier *ast.Ident,
) (exposed []ast.Expr) {
	isKnown := slices.Contains(
		predeclaredTypes(),
		identifier.Name,
	) || slices.Contains(
		found.names,
		identifier.Name,
	) || slices.Contains(
		typeParameters,
		identifier.Name,
	)
	if isKnown {
		return nil
	}

	return []ast.Expr{identifier}
}

func selectorExposures(
	found unit,
	selector *ast.SelectorExpr,
) (exposed []ast.Expr) {
	qualifier, isIdentifier := selector.X.(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	for _, imported := range importsNamed(
		found.file.Syntax,
		qualifier.Name,
	) {
		exposed = append(
			exposed,
			internalExposure(
				found,
				imported,
				selector,
			)...,
		)
	}

	return exposed
}

func internalExposure(
	found unit,
	imported string,
	selector *ast.SelectorExpr,
) (exposed []ast.Expr) {
	rel, inModule := strings.CutPrefix(
		imported,
		found.modulePath+pathSeparator,
	)
	isOwnInternal := inModule && strings.HasPrefix(
		rel+pathSeparator,
		internalPath(found.directory),
	)
	if !isOwnInternal {
		return nil
	}

	return []ast.Expr{selector}
}

func predeclaredTypes() (names []string) {
	return []string{"any", "bool", "byte", "comparable", "complex64", "complex128", "error", "float32", "float64", "int", "int8", "int16", "int32", "int64", "rune", "string", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr"}
}
