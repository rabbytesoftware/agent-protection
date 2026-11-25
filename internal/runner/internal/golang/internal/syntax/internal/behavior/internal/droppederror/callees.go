package droppederror

import (
	"go/ast"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	errorType          = "error"
	qualifierSeparator = "."
	importSeparator    = "/"
	blankImport        = "_"
	dotImport          = "."
	rootDirectory      = "."
	importQuotes       = "\"`"
	errorOnly          = 1
	valueAndError      = 2
	pairAndError       = 3
	tripleAndError     = 4
)

func erroringCallees(
	files []source.File,
) (callees map[string]map[string]int) {
	callees = map[string]map[string]int{}
	for _, file := range files {
		directory := filepath.Dir(file.Path)
		callees[directory] = merged(
			callees[directory],
			fileCallees(file),
		)
	}

	return callees
}

func merged(
	known map[string]int,
	found map[string]int,
) (union map[string]int) {
	union = map[string]int{}
	maps.Copy(
		union,
		known,
	)
	maps.Copy(
		union,
		found,
	)
	return union
}

func fileCallees(
	file source.File,
) (callees map[string]int) {
	callees = map[string]int{}
	for _, field := range source.Nodes[*ast.Field](file.Syntax) {
		maps.Copy(
			callees,
			fieldCallees(field),
		)
	}

	functions := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotErroringFunction,
	)
	for _, function := range functions {
		callees[function.Name.Name] = resultCount(function.Type)
	}

	return callees
}

func fieldCallees(
	field *ast.Field,
) (callees map[string]int) {
	signature, isFunc := field.Type.(*ast.FuncType)
	isErroring := isFunc && returnsError(signature)
	if !isErroring {
		return nil
	}

	callees = map[string]int{}
	for _, name := range field.Names {
		callees[name.Name] = resultCount(signature)
	}

	return callees
}

func isNotErroringFunction(
	function *ast.FuncDecl,
) (excluded bool) {
	return !returnsError(function.Type)
}

func returnsError(
	signature *ast.FuncType,
) (erroring bool) {
	if signature.Results == nil {
		return false
	}

	results := signature.Results.List
	return isErrorType(results[len(results)-1].Type)
}

func resultCount(
	signature *ast.FuncType,
) (count int) {
	for _, result := range signature.Results.List {
		count += max(
			1,
			len(result.Names),
		)
	}

	return count
}

func isErrorType(
	expression ast.Expr,
) (isError bool) {
	name, isIdent := expression.(*ast.Ident)
	return isIdent && name.Name == errorType
}

func knownCallees(
	file source.File,
	byDirectory map[string]map[string]int,
	directories map[string]string,
) (callees map[string]int) {
	callees = merged(
		byDirectory[filepath.Dir(file.Path)],
		nil,
	)
	for _, spec := range file.Syntax.Imports {
		maps.Copy(
			callees,
			importCallees(
				spec,
				directories,
				byDirectory,
			),
		)
	}

	return callees
}

func importCallees(
	spec *ast.ImportSpec,
	directories map[string]string,
	byDirectory map[string]map[string]int,
) (callees map[string]int) {
	importPath := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	return qualified(
		localName(
			spec,
			importPath,
		),
		importedCallees(
			importPath,
			directories,
			byDirectory,
		),
	)
}

func relativeDirectories(
	files []source.File,
) (directories map[string]string) {
	directories = map[string]string{}
	for _, file := range files {
		directories[filepath.ToSlash(filepath.Dir(file.Rel))] = filepath.Dir(file.Path)
	}

	return directories
}

func importedCallees(
	importPath string,
	directories map[string]string,
	byDirectory map[string]map[string]int,
) (callees map[string]int) {
	known, isStandard := standardCallees()[importPath]
	if isStandard {
		return known
	}

	matched := ""
	for slashed := range directories {
		matched = longerMatch(
			importPath,
			slashed,
			matched,
		)
	}

	return byDirectory[directories[matched]]
}

func longerMatch(
	importPath string,
	candidate string,
	matched string,
) (longest string) {
	isLonger := len(candidate) > len(matched) && candidate != rootDirectory && strings.HasSuffix(
		importSeparator+importPath,
		importSeparator+candidate,
	)
	if !isLonger {
		return matched
	}

	return candidate
}

func localName(
	spec *ast.ImportSpec,
	importPath string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(importPath)
}

func qualified(
	qualifier string,
	callees map[string]int,
) (named map[string]int) {
	named = map[string]int{}
	isUsable := qualifier != blankImport && qualifier != dotImport
	if !isUsable {
		return named
	}

	for name, count := range callees {
		named[qualifier+qualifierSeparator+name] = count
	}

	return named
}

func standardCallees() (callees map[string]map[string]int) {
	return map[string]map[string]int{
		"strconv":       strconvCallees(),
		"os":            osCallees(),
		"io":            ioCallees(),
		"encoding/json": jsonCallees(),
		"fmt":           fmtCallees(),
	}
}

func strconvCallees() (callees map[string]int) {
	return map[string]int{
		"Atoi":         valueAndError,
		"ParseBool":    valueAndError,
		"ParseFloat":   valueAndError,
		"ParseInt":     valueAndError,
		"ParseUint":    valueAndError,
		"ParseComplex": valueAndError,
		"Unquote":      valueAndError,
		"UnquoteChar":  tripleAndError,
		"QuotedPrefix": valueAndError,
	}
}

func osCallees() (callees map[string]int) {
	return merged(
		map[string]int{
			"Chdir":     errorOnly,
			"Chmod":     errorOnly,
			"Chown":     errorOnly,
			"Chtimes":   errorOnly,
			"CopyFS":    errorOnly,
			"Lchown":    errorOnly,
			"Link":      errorOnly,
			"Mkdir":     errorOnly,
			"MkdirAll":  errorOnly,
			"Remove":    errorOnly,
			"RemoveAll": errorOnly,
			"Rename":    errorOnly,
			"Setenv":    errorOnly,
			"Symlink":   errorOnly,
			"Truncate":  errorOnly,
			"Unsetenv":  errorOnly,
			"WriteFile": errorOnly,
		},
		osValueCallees(),
	)
}

func osValueCallees() (callees map[string]int) {
	return map[string]int{
		"Create":        valueAndError,
		"CreateTemp":    valueAndError,
		"Executable":    valueAndError,
		"FindProcess":   valueAndError,
		"Getwd":         valueAndError,
		"Hostname":      valueAndError,
		"Lstat":         valueAndError,
		"MkdirTemp":     valueAndError,
		"Open":          valueAndError,
		"OpenFile":      valueAndError,
		"Pipe":          pairAndError,
		"ReadDir":       valueAndError,
		"ReadFile":      valueAndError,
		"Readlink":      valueAndError,
		"Stat":          valueAndError,
		"UserCacheDir":  valueAndError,
		"UserConfigDir": valueAndError,
		"UserHomeDir":   valueAndError,
	}
}

func ioCallees() (callees map[string]int) {
	return map[string]int{
		"Copy":        valueAndError,
		"CopyBuffer":  valueAndError,
		"CopyN":       valueAndError,
		"ReadAll":     valueAndError,
		"ReadAtLeast": valueAndError,
		"ReadFull":    valueAndError,
		"WriteString": valueAndError,
	}
}

func jsonCallees() (callees map[string]int) {
	return map[string]int{
		"Compact":       errorOnly,
		"Indent":        errorOnly,
		"Marshal":       valueAndError,
		"MarshalIndent": valueAndError,
		"Unmarshal":     errorOnly,
	}
}

func fmtCallees() (callees map[string]int) {
	return map[string]int{
		"Fscan":   valueAndError,
		"Fscanf":  valueAndError,
		"Fscanln": valueAndError,
		"Scan":    valueAndError,
		"Scanf":   valueAndError,
		"Scanln":  valueAndError,
		"Sscan":   valueAndError,
		"Sscanf":  valueAndError,
		"Sscanln": valueAndError,
	}
}
