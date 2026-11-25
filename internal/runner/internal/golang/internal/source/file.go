package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

const (
	importQuotes  = "\"`"
	rootDirectory = "."
	dotImport     = "."
	testingImport = "testing"
)

type File struct {
	Path        string
	Rel         string
	FileSet     *token.FileSet
	Syntax      *ast.File
	Content     []byte
	IsGenerated bool
}

func (f File) Location(
	node ast.Node,
) (location string) {
	return fmt.Sprintf(
		"%s:%d",
		f.Rel,
		f.Line(node),
	)
}

func (f File) Line(
	node ast.Node,
) (line int) {
	return f.FileSet.Position(node.Pos()).Line
}

func (f File) Column(
	node ast.Node,
) (column int) {
	return f.FileSet.Position(node.Pos()).Column
}

func (f File) Offset(
	position token.Pos,
) (offset int) {
	return f.FileSet.Position(position).Offset
}

func (f File) LineOffset(
	line int,
) (offset int) {
	tokenFile := f.FileSet.File(f.Syntax.Pos())
	return tokenFile.Offset(tokenFile.LineStart(line))
}

func MockFiles(
	files []File,
) (mocks []string) {
	users := MockUsers(files)
	for _, file := range files {
		mocks = append(
			mocks,
			unusedMock(
				file,
				users,
			)...,
		)
	}

	return mocks
}

func WithoutMocks(
	files []File,
) (kept []File) {
	mocks := MockFiles(files)
	for _, file := range files {
		kept = append(
			kept,
			unlisted(
				file,
				mocks,
			)...,
		)
	}

	return kept
}

func MockUsers(
	files []File,
) (users map[string]string) {
	return withUsers(
		files,
		map[string]string{},
	)
}

func withUsers(
	files []File,
	known map[string]string,
) (users map[string]string) {
	found := map[string]string{}
	referrers := productionFiles(
		files,
		known,
	)
	for _, candidate := range pendingMocks(
		files,
		known,
	) {
		maps.Copy(
			found,
			userOf(
				candidate,
				referrers,
			),
		)
	}

	if len(found) == 0 {
		return known
	}

	users = maps.Clone(known)
	maps.Copy(
		users,
		found,
	)
	return withUsers(
		files,
		users,
	)
}

func unusedMock(
	file File,
	users map[string]string,
) (mocks []string) {
	_, isUsed := users[file.Rel]
	if isUsed ||
		!isMockSource(file) {
		return nil
	}

	return []string{file.Rel}
}

func unlisted(
	file File,
	listed []string,
) (kept []File) {
	if slices.Contains(
		listed,
		file.Rel,
	) {
		return nil
	}

	return []File{file}
}

func pendingMocks(
	files []File,
	users map[string]string,
) (pending []File) {
	for _, file := range files {
		pending = append(
			pending,
			pendingMock(
				file,
				users,
			)...,
		)
	}

	return pending
}

func pendingMock(
	file File,
	users map[string]string,
) (pending []File) {
	_, isUsed := users[file.Rel]
	if isUsed ||
		!isMockSource(file) {
		return nil
	}

	return []File{file}
}

func productionFiles(
	files []File,
	users map[string]string,
) (production []File) {
	for _, file := range files {
		production = append(
			production,
			productionFile(
				file,
				users,
			)...,
		)
	}

	return production
}

func productionFile(
	file File,
	users map[string]string,
) (production []File) {
	_, isUsedMock := users[file.Rel]
	isProduction := !IsTest(file.Rel) && !importsTesting(file.Syntax) && (isUsedMock || !IsMock(file.Rel))
	if !isProduction {
		return nil
	}

	return []File{file}
}

func userOf(
	candidate File,
	referrers []File,
) (user map[string]string) {
	users := []string{}
	for _, referrer := range referrers {
		users = append(
			users,
			referencing(
				referrer,
				candidate,
			)...,
		)
	}

	if len(users) == 0 {
		return nil
	}

	return map[string]string{candidate.Rel: users[0]}
}

func referencing(
	referrer File,
	candidate File,
) (rels []string) {
	if !references(
		referrer,
		candidate,
	) {
		return nil
	}

	return []string{referrer.Rel}
}

func references(
	referrer File,
	candidate File,
) (referenced bool) {
	directory := filepath.ToSlash(filepath.Dir(candidate.Rel))
	if filepath.ToSlash(filepath.Dir(referrer.Rel)) == directory {
		return referrer.Syntax.Name.Name == candidate.Syntax.Name.Name && (usesAny(
			referrer.Syntax.Unresolved,
			declaredNames(candidate.Syntax),
		) || usesAny(
			selectedNames(referrer.Syntax),
			foreignMethods(candidate.Syntax),
		))
	}

	for _, spec := range referrer.Syntax.Imports {
		referenced = referenced || importUses(
			referrer.Syntax,
			spec,
			candidate,
		)
	}

	return referenced
}

func importUses(
	syntax *ast.File,
	spec *ast.ImportSpec,
	candidate File,
) (uses bool) {
	directory := filepath.ToSlash(filepath.Dir(candidate.Rel))
	imported := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	isImported := directory != rootDirectory && (imported == directory || strings.HasSuffix(
		imported,
		"/"+directory,
	))
	if !isImported {
		return false
	}

	exported := slices.DeleteFunc(
		declaredNames(candidate.Syntax),
		isUnexported,
	)
	qualifier := importQualifier(
		spec,
		candidate.Syntax.Name.Name,
	)
	if qualifier == dotImport {
		return usesAny(
			syntax.Unresolved,
			exported,
		)
	}

	return usesAny(
		qualifiedNames(
			syntax,
			qualifier,
		),
		exported,
	)
}

func importQualifier(
	spec *ast.ImportSpec,
	packageName string,
) (qualifier string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return packageName
}

func importsTesting(
	syntax *ast.File,
) (imports bool) {
	for _, spec := range syntax.Imports {
		imports = imports || strings.Trim(
			spec.Path.Value,
			importQuotes,
		) == testingImport
	}

	return imports
}

func isMockSource(
	file File,
) (isMock bool) {
	return !IsTest(file.Rel) && IsMock(file.Rel)
}

func isUnexported(
	name string,
) (excluded bool) {
	return !ast.IsExported(name)
}
