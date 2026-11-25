package publicapi

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	containerFile = "container.go"
	apiFormat     = "%s %s in %s: %s"
)

func apiFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	if !found.api {
		return nil
	}

	function, isFunction := declaration.(*ast.FuncDecl)
	isEntry := isFunction && isEntryFunction(
		file,
		function,
	)
	if isEntry {
		return nil
	}

	return slices.Concat(
		violationFindings(
			file,
			found,
			declaration,
		),
		aliasFindings(
			file,
			found,
			declaration,
		),
	)
}

func violationFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, flagged := range violations(declaration) {
		findings = append(
			findings,
			report.Finding{
				Rule:     apiRule,
				Location: file.Location(flagged.node),
				Snippet: fmt.Sprintf(
					apiFormat,
					flagged.kind,
					flagged.name,
					found.label,
					apiRemedy(
						flagged,
						found,
					),
				),
			},
		)
	}

	return findings
}

func isEntryFunction(
	file source.File,
	function *ast.FuncDecl,
) (isEntry bool) {
	return filepath.Base(file.Path) == containerFile && function.Recv == nil && function.Name.IsExported()
}

func apiRemedy(
	flagged violation,
	found scope,
) (remedy string) {
	switch flagged.kind {
	case methodKind:
		return fmt.Sprintf(
			"implementation lives under %s; expose an interface here and construct the implementation in container.go",
			found.internal,
		)
	case funcKind:
		return functionRemedy(
			flagged,
			found,
		)
	case errorMethodKind:
		return fmt.Sprintf(
			"Error, Unwrap, Is, As and String hold a single return of a field, a literal, a concatenation or one fmt.Sprintf, fmt.Errorf or errors call; move the logic under %s",
			found.internal,
		)
	}

	return fmt.Sprintf(
		"the API holds only exported types, constants and Err sentinels; move it under %s",
		found.internal,
	)
}

func functionRemedy(
	flagged violation,
	found scope,
) (remedy string) {
	if flagged.exported {
		return fmt.Sprintf(
			"entry functions live in container.go and only delegate; move it there or move its logic under %s",
			found.internal,
		)
	}

	return fmt.Sprintf(
		"move it under %s",
		found.internal,
	)
}
