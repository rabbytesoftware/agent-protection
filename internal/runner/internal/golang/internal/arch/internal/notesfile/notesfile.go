package notesfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R07"
	snippetFormat = "%s is a notes file beside the code: what the package does belongs in its names and structure, and why it changed in the commit message; delete it (a README.md at the module root or a docs/ tree is where documentation lives, and a file the package embeds with //go:embed is data)"
	goExtension   = ".go"
	embedPattern  = `(?m)^//go:embed\s+(.+)$`
	patternGroup  = 1
	allMatches    = -1
	fieldPattern  = `"(?:[^"\\]|\\.)*"|` + "`[^`]*`" + `|\S+`
	allPrefix     = "all:"
	quoteRunes    = "\"`"
	quotedLength  = 2
	pathSeparator = "/"
	rootDirectory = "."
)

func Notesfile(
	root string,
) (findings []report.Finding, err error) {
	directories, err := source.Directories(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"notes files: %w",
				err,
			)
	}

	failures := []error{}
	for _, directory := range directories {
		found, directoryErr := directoryFindings(
			root,
			directory,
		)
		findings = append(
			findings,
			found...,
		)
		failures = append(
			failures,
			directoryErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func directoryFindings(
	root string,
	directory string,
) (findings []report.Finding, err error) {
	rel, err := source.RelPath(
		root,
		directory,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"notes files: %w",
				err,
			)
	}

	if isDocumentation(rel) {
		return nil,
			nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil,
			fmt.Errorf(
				"notes files: read %s: %w",
				directory,
				err,
			)
	}

	return notesFindings(
		directory,
		rel,
		fileNames(entries),
	)
}

func notesFindings(
	directory string,
	rel string,
	names []string,
) (findings []report.Finding, err error) {
	sources := slices.DeleteFunc(
		slices.Clone(names),
		isNotGoFile,
	)
	if len(sources) == 0 {
		return nil,
			nil
	}

	patterns, err := embedPatterns(
		directory,
		sources,
	)
	found, notesErr := namedFindings(
		rel,
		slices.DeleteFunc(
			slices.Clone(names),
			isNotNotes,
		),
		patterns,
	)
	return found,
		errors.Join(
			err,
			notesErr,
		)
}

func namedFindings(
	rel string,
	names []string,
	patterns []string,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, name := range names {
		found, noteErr := noteFinding(
			filepath.ToSlash(filepath.Join(
				rel,
				name,
			)),
			name,
			patterns,
		)
		findings = append(
			findings,
			found...,
		)
		failures = append(
			failures,
			noteErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func noteFinding(
	location string,
	name string,
	patterns []string,
) (findings []report.Finding, err error) {
	embedded, err := isEmbedded(
		name,
		patterns,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"notes files: embed pattern beside %s: %w",
				location,
				err,
			)
	}

	if embedded {
		return nil,
			nil
	}

	return []report.Finding{{
			Rule:     rule,
			Location: location,
			Snippet: fmt.Sprintf(
				snippetFormat,
				location,
			),
		}},
		nil
}

func embedPatterns(
	directory string,
	sources []string,
) (patterns []string, err error) {
	failures := []error{}
	for _, name := range sources {
		content, readErr := os.ReadFile(filepath.Join(
			directory,
			name,
		))
		patterns = append(
			patterns,
			directivePatterns(string(content))...,
		)
		failures = append(
			failures,
			readErr,
		)
	}

	return patterns,
		errors.Join(failures...)
}

func directivePatterns(
	content string,
) (patterns []string) {
	for _, match := range regexp.MustCompile(embedPattern).FindAllStringSubmatch(
		content,
		allMatches,
	) {
		patterns = append(
			patterns,
			unquoted(regexp.MustCompile(fieldPattern).FindAllString(
				match[patternGroup],
				allMatches,
			))...,
		)
	}

	return patterns
}

func unquoted(
	fields []string,
) (patterns []string) {
	for _, field := range fields {
		patterns = append(
			patterns,
			unquotedField(field),
		)
	}

	return patterns
}

func unquotedField(
	field string,
) (pattern string) {
	isQuoted := len(field) >= quotedLength && strings.ContainsRune(
		quoteRunes,
		rune(field[0]),
	)
	if !isQuoted {
		return field
	}

	return field[1 : len(field)-1]
}

func isEmbedded(
	name string,
	patterns []string,
) (embedded bool, err error) {
	failures := []error{}
	for _, pattern := range patterns {
		matched, matchErr := filepath.Match(
			strings.TrimPrefix(
				pattern,
				allPrefix,
			),
			name,
		)
		embedded = embedded || matched
		failures = append(
			failures,
			matchErr,
		)
	}

	return embedded,
		errors.Join(failures...)
}

func fileNames(
	entries []os.DirEntry,
) (names []string) {
	for _, entry := range slices.DeleteFunc(
		entries,
		isDirectory,
	) {
		names = append(
			names,
			entry.Name(),
		)
	}

	return names
}

func isDirectory(
	entry os.DirEntry,
) (directory bool) {
	return entry.IsDir()
}

func isNotGoFile(
	name string,
) (excluded bool) {
	return filepath.Ext(name) != goExtension
}

func isNotNotes(
	name string,
) (excluded bool) {
	return !slices.Contains(
		notesExtensions(),
		strings.ToLower(filepath.Ext(name)),
	)
}

func isDocumentation(
	rel string,
) (documentation bool) {
	segments := strings.Split(
		filepath.ToSlash(rel),
		pathSeparator,
	)
	return rel == rootDirectory || source.IsExample(rel) || slices.ContainsFunc(
		segments,
		isDocumentationSegment,
	)
}

func isDocumentationSegment(
	segment string,
) (documentation bool) {
	return slices.Contains(
		documentationDirectories(),
		segment,
	)
}

func notesExtensions() (extensions []string) {
	return []string{
		".md",
		".txt",
		".rst",
	}
}

func documentationDirectories() (names []string) {
	return []string{
		"docs",
		"doc",
		"testdata",
	}
}
