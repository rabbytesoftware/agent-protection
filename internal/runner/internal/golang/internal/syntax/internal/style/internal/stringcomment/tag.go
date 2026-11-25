package stringcomment

import (
	_ "embed"
	"fmt"
	"go/ast"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	keySnippetFormat = "struct tag key %s is not read by any known library: a tag carries machine metadata, not explanations; delete it and make the code say it through names and structure"
	malformedSnippet = "struct tag is not a list of key:\"value\" pairs: a tag carries machine metadata, not explanations; delete the prose"
	tagPattern       = `^\s*(?:[^\s:"\x7f]+:"(?:[^"\\]|\\.)*"\s*)*$`
	keyPattern       = `([^\s:"\x7f]+):"(?:[^"\\]|\\.)*"`
	keyGroup         = 1
	allMatches       = -1
)

func tagFindings(
	file source.File,
) (findings []report.Finding) {
	for _, field := range source.Nodes[*ast.Field](file.Syntax) {
		findings = append(
			findings,
			fieldFindings(
				file,
				field,
			)...,
		)
	}

	return findings
}

func fieldFindings(
	file source.File,
	field *ast.Field,
) (findings []report.Finding) {
	if field.Tag == nil {
		return nil
	}

	location := file.Location(field.Tag)
	tag, wellFormed := tagValue(field.Tag.Value)
	if !wellFormed {
		return []report.Finding{{Rule: rule, Location: location, Snippet: malformedSnippet}}
	}

	for _, key := range unknownKeys(tag) {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: location,
				Snippet: fmt.Sprintf(
					keySnippetFormat,
					strconv.Quote(key),
				),
			},
		)
	}

	return findings
}

func tagValue(
	literal string,
) (tag string, wellFormed bool) {
	tag, err := strconv.Unquote(literal)
	return tag,
		err == nil && regexp.MustCompile(tagPattern).MatchString(tag)
}

func unknownKeys(
	tag string,
) (keys []string) {
	for _, match := range regexp.MustCompile(keyPattern).FindAllStringSubmatch(
		tag,
		allMatches,
	) {
		keys = append(
			keys,
			unknownKey(match[keyGroup])...,
		)
	}

	return keys
}

func unknownKey(
	key string,
) (unknown []string) {
	if slices.Contains(
		machineKeys(),
		key,
	) {
		return nil
	}

	return []string{key}
}

//go:embed machine_keys.txt
var machineKeyList string

func machineKeys() (keys []string) {
	return strings.Fields(machineKeyList)
}
