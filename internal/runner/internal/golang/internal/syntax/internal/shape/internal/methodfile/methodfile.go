package methodfile

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R66"
	numberedFormat  = "file %s has a numbered name: name the file after the one type it holds (or the package's entry function) and merge or split by type, not by part"
	misplacedFormat = "methods of %s live in %s: move them into %s, the file that declares %s"
)

func Methodfile(
	files []source.File,
) (findings []report.Finding) {
	checked := source.WithoutMocks(files)
	directories := map[string][]string{}
	for _, file := range checked {
		directory := filepath.Dir(file.Path)
		directories[directory] = append(
			directories[directory],
			portable(stem(file.Path)),
		)
	}

	for _, file := range checked {
		findings = slices.Concat(
			findings,
			numberedFindings(
				file,
				directories[filepath.Dir(file.Path)],
			),
			misplacedFindings(file),
		)
	}

	return findings
}

func numberedFindings(
	file source.File,
	siblings []string,
) (findings []report.Finding) {
	if !isNumbered(
		portable(stem(file.Path)),
		siblings,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Rel,
		Snippet: fmt.Sprintf(
			numberedFormat,
			filepath.Base(file.Path),
		),
	}}
}

func misplacedFindings(
	file source.File,
) (findings []report.Finding) {
	if source.IsCompositionRoot(
		file.Path,
		file.Syntax.Name.Name,
	) {
		return nil
	}

	for _, method := range firstMethods(file.Syntax) {
		findings = appendMisplaced(
			findings,
			file,
			method,
		)
	}

	return findings
}

func appendMisplaced(
	findings []report.Finding,
	file source.File,
	method *ast.FuncDecl,
) (extended []report.Finding) {
	receiver := receiverType(method)
	if snakeCase(receiver) == portable(stem(file.Path)) {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(method),
			Snippet: fmt.Sprintf(
				misplacedFormat,
				receiver,
				filepath.Base(file.Path),
				snakeCase(receiver)+goExtension,
				receiver,
			),
		},
	)
}

func firstMethods(
	syntax *ast.File,
) (first []*ast.FuncDecl) {
	for _, method := range methods(syntax) {
		first = appendFirst(
			first,
			method,
		)
	}

	return first
}

func appendFirst(
	first []*ast.FuncDecl,
	method *ast.FuncDecl,
) (extended []*ast.FuncDecl) {
	isRepeat := slices.ContainsFunc(
		first,
		sameReceiver(method),
	)
	if isRepeat {
		return first
	}

	return append(
		first,
		method,
	)
}

func methods(
	syntax *ast.File,
) (declarations []*ast.FuncDecl) {
	for _, declaration := range syntax.Decls {
		declarations = appendMethod(
			declarations,
			declaration,
		)
	}

	return declarations
}

func appendMethod(
	declarations []*ast.FuncDecl,
	declaration ast.Decl,
) (extended []*ast.FuncDecl) {
	function, isFunction := declaration.(*ast.FuncDecl)
	isMethod := isFunction &&
		function.Recv != nil &&
		len(function.Recv.List) > 0
	if !isMethod {
		return declarations
	}

	return append(
		declarations,
		function,
	)
}

func receiverType(
	method *ast.FuncDecl,
) (name string) {
	return typeName(method.Recv.List[0].Type)
}

func typeName(
	expression ast.Expr,
) (name string) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return typeName(typed.X)
	case *ast.IndexExpr:
		return typeName(typed.X)
	case *ast.IndexListExpr:
		return typeName(typed.X)
	case *ast.ParenExpr:
		return typeName(typed.X)
	}

	return ""
}

func sameReceiver(
	method *ast.FuncDecl,
) (matches func(*ast.FuncDecl) bool) {
	return func(
		chosen *ast.FuncDecl,
	) (same bool) {
		return receiverType(chosen) == receiverType(method)
	}
}
