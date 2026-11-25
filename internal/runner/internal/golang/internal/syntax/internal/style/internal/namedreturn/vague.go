package namedreturn

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	vagueSnippet = "%s gives results vague names (%s): name every result after what it holds, for example (%s)"
	opaqueName   = "value"
	countName    = "n"
	digits       = "0123456789"
	anyType      = "any"
)

func vagueFindings(
	file source.File,
) (findings []report.Finding) {
	declarations := source.Nodes[*ast.FuncDecl](file.Syntax)
	for _, function := range declarations {
		findings = append(
			findings,
			vagueResultFindings(
				file,
				declarationSubject(function),
				function.Type,
				function.Body,
				append(
					typeParameters(function),
					function.Name.Name,
				),
			)...,
		)
	}

	for _, literal := range source.Nodes[*ast.FuncLit](file.Syntax) {
		findings = append(
			findings,
			vagueResultFindings(
				file,
				literalSubject,
				literal.Type,
				literal.Body,
				enclosingParameters(
					declarations,
					literal,
				),
			)...,
		)
	}

	return findings
}

func enclosingParameters(
	declarations []*ast.FuncDecl,
	literal *ast.FuncLit,
) (names []string) {
	for _, function := range declarations {
		names = append(
			names,
			parametersAround(
				function,
				literal,
			)...,
		)
	}

	return names
}

func parametersAround(
	function *ast.FuncDecl,
	literal *ast.FuncLit,
) (names []string) {
	isAround := function.Pos() <= literal.Pos() && literal.End() <= function.End()
	if !isAround {
		return nil
	}

	return typeParameters(function)
}

func vagueResultFindings(
	file source.File,
	subject string,
	signature *ast.FuncType,
	body *ast.BlockStmt,
	scope []string,
) (findings []report.Finding) {
	isChecked := body != nil && signature.Results != nil
	if !isChecked {
		return nil
	}

	vague := vagueNames(
		signature.Results,
		scope,
	)
	if len(vague) == 0 {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(vague[0]),
		Snippet: fmt.Sprintf(
			vagueSnippet,
			subject,
			vagueListing(
				signature.Results,
				scope,
			),
			renamedResults(
				signature.Results,
				scope,
			),
		),
	}}
}

func vagueNames(
	results *ast.FieldList,
	scope []string,
) (names []*ast.Ident) {
	names, kinds := flattened(results)
	for index, name := range names {
		names[index] = vagueIdent(
			name,
			kinds[index],
			scope,
		)
	}

	return slices.DeleteFunc(
		names,
		isNilIdent,
	)
}

func vagueIdent(
	name *ast.Ident,
	kind ast.Expr,
	scope []string,
) (vague *ast.Ident) {
	if !isVague(
		name.Name,
		kind,
		scope,
	) {
		return nil
	}

	return name
}

func isNilIdent(
	name *ast.Ident,
) (absent bool) {
	return name == nil
}

func flattened(
	results *ast.FieldList,
) (names []*ast.Ident, kinds []ast.Expr) {
	for _, field := range results.List {
		names = append(
			names,
			field.Names...,
		)
		kinds = append(
			kinds,
			slices.Repeat(
				[]ast.Expr{field.Type},
				len(field.Names),
			)...,
		)
	}

	return names,
		kinds
}

func vagueListing(
	results *ast.FieldList,
	scope []string,
) (listed string) {
	names, kinds := flattened(results)
	written := []string{}
	for index, name := range names {
		written = appendListed(
			written,
			name,
			kinds[index],
			scope,
		)
	}

	return strings.Join(
		written,
		listSeparator,
	)
}

func appendListed(
	written []string,
	name *ast.Ident,
	kind ast.Expr,
	scope []string,
) (extended []string) {
	if !isVague(
		name.Name,
		kind,
		scope,
	) {
		return written
	}

	return append(
		written,
		name.Name+" "+types.ExprString(kind),
	)
}

func renamedResults(
	results *ast.FieldList,
	scope []string,
) (listed string) {
	names, kinds := flattened(results)
	written := []string{}
	for index, name := range names {
		written = append(
			written,
			renamed(
				name.Name,
				kinds[index],
				scope,
			)+" "+types.ExprString(kinds[index]),
		)
	}

	return strings.Join(
		written,
		listSeparator,
	)
}

func renamed(
	name string,
	kind ast.Expr,
	scope []string,
) (replacement string) {
	if !isVague(
		name,
		kind,
		scope,
	) {
		return name
	}

	if isOpaque(
		kind,
		scope,
	) {
		return opaqueName
	}

	return suggestedName(kind)
}

func isVague(
	name string,
	kind ast.Expr,
	scope []string,
) (vague bool) {
	if isNumberedVague(name) ||
		isLetter(name) {
		return true
	}

	return isVagueWord(name) && !isExcused(
		name,
		kind,
		scope,
	)
}

func isNumberedVague(
	name string,
) (numbered bool) {
	stem := strings.TrimRight(
		name,
		digits,
	)
	return stem != name && (isVagueWord(stem) || slices.Contains(
		meaningfulShortNames(),
		stem,
	))
}

func isExcused(
	name string,
	kind ast.Expr,
	scope []string,
) (excused bool) {
	return isTypeWord(
		name,
		kind,
	) || isOwnName(
		name,
		scope,
	) || name == opaqueName && isOpaque(
		kind,
		scope,
	)
}

func isOwnName(
	name string,
	scope []string,
) (own bool) {
	for _, declared := range scope {
		own = own || strings.EqualFold(
			declared,
			name,
		)
	}

	return own
}

func isVagueWord(
	name string,
) (vague bool) {
	return isLetter(name) || slices.Contains(
		vagueWords(),
		name,
	)
}

func isLetter(
	name string,
) (letter bool) {
	runes := []rune(name)
	return name != countName &&
		len(runes) == 1 &&
		(unicode.IsLetter(runes[0]) || runes[0] == '_')
}

func isTypeWord(
	name string,
	kind ast.Expr,
) (matches bool) {
	written := rawTypeName(kind)
	titled := strings.ToUpper(name[:1]) + name[1:]
	return strings.EqualFold(
		written,
		name,
	) || strings.HasSuffix(
		written,
		titled,
	)
}

func rawTypeName(
	kind ast.Expr,
) (name string) {
	switch typed := kind.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return rawTypeName(typed.X)
	case *ast.ArrayType:
		return pluralName(rawTypeName(typed.Elt))
	case *ast.IndexExpr:
		return rawTypeName(typed.X)
	case *ast.IndexListExpr:
		return rawTypeName(typed.X)
	}

	return ""
}

func isOpaque(
	kind ast.Expr,
	scope []string,
) (opaque bool) {
	switch typed := kind.(type) {
	case *ast.Ident:
		return typed.Name == anyType || slices.Contains(
			scope,
			typed.Name,
		)
	case *ast.InterfaceType:
		return typed.Methods == nil || len(typed.Methods.List) == 0
	}

	return false
}

func typeParameters(
	function *ast.FuncDecl,
) (names []string) {
	names = fieldNames(function.Type.TypeParams)
	if function.Recv == nil {
		return names
	}

	for _, field := range function.Recv.List {
		names = append(
			names,
			receiverParameters(field.Type)...,
		)
	}

	return names
}

func receiverParameters(
	kind ast.Expr,
) (names []string) {
	switch typed := kind.(type) {
	case *ast.StarExpr:
		return receiverParameters(typed.X)
	case *ast.IndexExpr:
		return identNames([]ast.Expr{typed.Index})
	case *ast.IndexListExpr:
		return identNames(typed.Indices)
	}

	return nil
}

func fieldNames(
	fields *ast.FieldList,
) (names []string) {
	if fields == nil {
		return nil
	}

	for _, field := range fields.List {
		names = append(
			names,
			identNames(identExpressions(field.Names))...,
		)
	}

	return names
}

func identExpressions(
	idents []*ast.Ident,
) (expressions []ast.Expr) {
	for _, ident := range idents {
		expressions = append(
			expressions,
			ident,
		)
	}

	return expressions
}

func meaningfulShortNames() (names []string) {
	return []string{
		errorName,
		"ok",
		countName,
	}
}

func vagueWords() (words []string) {
	return []string{
		"res",
		"ret",
		"rv",
		"rval",
		"retval",
		"val",
		"value",
		"out",
		"output",
		"tmp",
		"temp",
		"data",
		"result",
		"results",
		"ans",
	}
}
