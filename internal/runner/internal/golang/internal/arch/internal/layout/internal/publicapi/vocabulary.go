package publicapi

import (
	"fmt"
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	vocabularyFormat  = "%s %s in vocabulary package %s: %s"
	behaviorRemedy    = "vocabulary packages hold declarations only; move behavior into the package that uses it"
	declarationRemedy = "vocabulary packages hold only exported types, constants and Err sentinels; move it into the package that uses it"
)

func vocabularyFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, flagged := range violations(declaration) {
		findings = append(
			findings,
			report.Finding{
				Rule:     vocabularyRule,
				Location: file.Location(flagged.node),
				Snippet: fmt.Sprintf(
					vocabularyFormat,
					flagged.kind,
					flagged.name,
					found.packageName,
					vocabularyRemedy(flagged),
				),
			},
		)
	}

	return findings
}

func vocabularyRemedy(
	flagged violation,
) (remedy string) {
	isBehavior := flagged.kind == funcKind || flagged.kind == methodKind || flagged.kind == errorMethodKind
	if isBehavior {
		return behaviorRemedy
	}

	return declarationRemedy
}

func vocabularyNames() (names []string) {
	return []string{
		"models",
		"model",
		"contracts",
		"contract",
		"types",
		"dto",
		"ports",
		"api",
		"options",
		"entities",
		"schema",
	}
}
