package interfacesize

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R28"
	maxMethods    = 3
	listSeparator = ", "
	findingFormat = "interface %s has %d methods once embedded %s are flattened (max %d): split it along what each consumer calls and have consumers depend on the small interfaces"
)

func Interfacesize(
	files []source.File,
) (findings []report.Finding) {
	checked := slices.DeleteFunc(
		slices.Clone(files),
		isExcluded,
	)
	index := declarations(checked)
	for _, file := range checked {
		findings = append(
			findings,
			fileFindings(
				file,
				index,
			)...,
		)
	}

	return findings
}

func isExcluded(
	file source.File,
) (excluded bool) {
	return source.IsTest(file.Path) || source.IsMock(file.Rel)
}

func fileFindings(
	file source.File,
	index map[string]declaration,
) (findings []report.Finding) {
	for _, spec := range interfaceSpecs(file.Syntax) {
		findings = append(
			findings,
			specFindings(
				file,
				index[declarationKey(
					packageDirectory(file),
					spec.Name.Name,
				)],
				index,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	declared declaration,
	index map[string]declaration,
) (findings []report.Finding) {
	embeds := resolvedEmbeds(
		declared,
		index,
	)
	methods := methodNames(
		declared,
		index,
		nil,
	)
	isBloated := len(embeds) > 0 &&
		len(declared.body.Methods.List) <= maxMethods &&
		len(methods) > maxMethods
	if !isBloated {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(declared.spec),
		Snippet: fmt.Sprintf(
			findingFormat,
			declared.spec.Name.Name,
			len(methods),
			embedNames(embeds),
			maxMethods,
		),
	}}
}

func embedNames(
	embeds []*ast.Field,
) (joined string) {
	names := []string{}
	for _, field := range embeds {
		names = append(
			names,
			types.ExprString(field.Type),
		)
	}

	return strings.Join(
		names,
		listSeparator,
	)
}
