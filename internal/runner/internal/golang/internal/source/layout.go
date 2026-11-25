package source

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	internalDirectory = "internal"
	moduleDirective   = "module "
	containerFile     = "container.go"
	mainFile          = "main.go"
	mainPackage       = "main"
)

const CompositionRootThreshold = 2

func ModulePath(
	root string,
) (modulePath string, err error) {
	content, err := os.ReadFile(filepath.Join(
		root,
		"go.mod",
	))
	if err != nil {
		return "",
			fmt.Errorf(
				"read go.mod: %w",
				err,
			)
	}

	lines := strings.Split(
		string(content),
		"\n",
	)
	index := slices.IndexFunc(
		lines,
		isModuleDirective,
	)
	if index < 0 {
		return "",
			fmt.Errorf(
				"no module directive in %s",
				filepath.Join(
					root,
					"go.mod",
				),
			)
	}

	return strings.TrimSpace(strings.TrimPrefix(
			strings.TrimSpace(lines[index]),
			moduleDirective,
		)),
		nil
}

func IsDeclarationOnly(
	path string,
) (declarationOnly bool, err error) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.SkipObjectResolution,
	)
	if err != nil {
		return false,
			fmt.Errorf(
				"parse %s: %w",
				path,
				err,
			)
	}

	return !slices.ContainsFunc(
			syntax.Decls,
			isFuncDecl,
		),
		nil
}

func IsInternalTree(
	dir string,
) (inside bool) {
	return slices.Contains(
		strings.Split(
			filepath.ToSlash(dir),
			"/",
		),
		internalDirectory,
	)
}

func IsInternalDirectory(
	dir string,
) (isInternal bool) {
	return filepath.Base(dir) == internalDirectory
}

func IsMock(
	rel string,
) (isMock bool) {
	segments := strings.Split(
		filepath.ToSlash(rel),
		"/",
	)
	last := len(segments) - 1
	return slices.ContainsFunc(
		segments[:last],
		isMockDirectory,
	) || isMockFileName(strings.TrimSuffix(
		segments[last],
		goExtension,
	))
}

func IsExample(
	rel string,
) (isExample bool) {
	return slices.ContainsFunc(
		strings.Split(
			filepath.ToSlash(rel),
			"/",
		),
		isExampleDirectory,
	)
}

func IsModule(
	dir string,
	sourceDirectories []string,
) (isModule bool) {
	prefix := filepath.Join(
		dir,
		internalDirectory,
	) + string(filepath.Separator)
	for _, sourceDirectory := range sourceDirectories {
		isModule = isModule || strings.HasPrefix(
			sourceDirectory,
			prefix,
		)
	}

	return isModule
}

func IsCompositionRoot(
	path string,
	packageName string,
) (isRoot bool) {
	base := filepath.Base(path)
	return base == containerFile || base == mainFile && packageName == mainPackage
}

func HasSource(
	dir string,
) (hasSource bool, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false,
			fmt.Errorf(
				"read %s: %w",
				dir,
				err,
			)
	}

	return slices.ContainsFunc(
			entries,
			isSourceEntry,
		),
		nil
}

func Entries(
	dir string,
) (paths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				dir,
				err,
			)
	}

	for _, entry := range slices.DeleteFunc(
		entries,
		isNotSourceEntry,
	) {
		paths = append(
			paths,
			filepath.Join(
				dir,
				entry.Name(),
			),
		)
	}

	return paths,
		nil
}

func isModuleDirective(
	line string,
) (matches bool) {
	return strings.HasPrefix(
		strings.TrimSpace(line),
		moduleDirective,
	)
}

func isFuncDecl(
	decl ast.Decl,
) (isFunc bool) {
	_, isFunc = decl.(*ast.FuncDecl)
	return isFunc
}

func isSourceEntry(
	entry os.DirEntry,
) (isSource bool) {
	return !entry.IsDir() && filepath.Ext(entry.Name()) == goExtension && !IsTest(entry.Name())
}

func isNotSourceEntry(
	entry os.DirEntry,
) (excluded bool) {
	return !isSourceEntry(entry)
}

func isMockDirectory(
	segment string,
) (matches bool) {
	return slices.Contains(
		[]string{"mock", "mocks", "fake", "fakes"},
		segment,
	)
}

func isExampleDirectory(
	segment string,
) (matches bool) {
	return slices.Contains(
		[]string{"example", "examples", "_examples"},
		segment,
	)
}

func isMockFileName(
	stem string,
) (matches bool) {
	return isMockDirectory(stem) ||
		strings.HasPrefix(
			stem,
			"mock_",
		) ||
		strings.HasPrefix(
			stem,
			"fake_",
		) ||
		strings.HasSuffix(
			stem,
			"_mock",
		) ||
		strings.HasSuffix(
			stem,
			"_fake",
		)
}

func hasIntegrationTag(
	comments []*ast.Comment,
) (tagged bool, err error) {
	failures := []error{}
	for _, comment := range slices.DeleteFunc(
		slices.Clone(comments),
		isNotGoBuild,
	) {
		integrationOnly, parseErr := isIntegrationOnly(comment.Text)
		tagged = tagged || integrationOnly
		failures = append(
			failures,
			parseErr,
		)
	}

	return tagged,
		errors.Join(failures...)
}

func isNotGoBuild(
	comment *ast.Comment,
) (excluded bool) {
	return !constraint.IsGoBuild(comment.Text)
}

func isIntegrationOnly(
	line string,
) (integrationOnly bool, err error) {
	expression, err := constraint.Parse(line)
	if err != nil {
		return false,
			fmt.Errorf(
				"parse build constraint %q: %w",
				line,
				err,
			)
	}

	others := slices.DeleteFunc(
		constraintTags(line),
		isIntegrationTag,
	)
	isEnumerable := len(others) <= maxEnumeratedTags
	return isEnumerable && isSatisfiable(
			expression,
			others,
			[]string{integrationTag},
		) && !isSatisfiable(
			expression,
			others,
			nil,
		),
		nil
}

func isSatisfiable(
	expression constraint.Expr,
	undecided []string,
	enabled []string,
) (satisfiable bool) {
	if len(undecided) == 0 {
		return expression.Eval(tagSetter(enabled))
	}

	return isSatisfiable(
		expression,
		undecided[1:],
		slices.Concat(
			enabled,
			undecided[:1],
		),
	) || isSatisfiable(
		expression,
		undecided[1:],
		enabled,
	)
}

func tagSetter(
	enabled []string,
) (isSet func(string) bool) {
	return func(
		tag string,
	) (set bool) {
		return slices.Contains(
			enabled,
			tag,
		)
	}
}

func constraintTags(
	line string,
) (tags []string) {
	tags = strings.FieldsFunc(
		strings.TrimPrefix(
			line,
			buildDirective,
		),
		isNotTagRune,
	)
	slices.Sort(tags)
	return slices.Compact(tags)
}

func isNotTagRune(
	character rune,
) (excluded bool) {
	isTagRune := unicode.IsLetter(character) || unicode.IsDigit(character) || character == '_' || character == '.'
	return !isTagRune
}

func isIntegrationTag(
	tag string,
) (matches bool) {
	return tag == integrationTag
}
