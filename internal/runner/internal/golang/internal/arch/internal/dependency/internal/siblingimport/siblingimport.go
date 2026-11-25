package siblingimport

import (
	"fmt"
	"go/ast"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	chainRule       = "R16"
	wiringRule      = "R47"
	moduleRoot      = "."
	containerFile   = "container.go"
	nameSeparator   = ", "
	importQuotes    = "\"`"
	wiringThreshold = 2
)

func Siblingimport(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"sibling imports: %w",
				err,
			)
	}

	files = slices.DeleteFunc(
		slices.Clone(files),
		isMock,
	)
	lifted, direct := edges(
		modulePath,
		files,
	)
	for _, file := range files {
		findings = append(
			findings,
			chainFindings(
				modulePath,
				lifted,
				file,
			)...,
		)
	}

	return slices.Concat(
			findings,
			wiringFindings(withLogic(
				direct,
				logicDirectories(files),
			)),
		),
		nil
}

func edges(
	modulePath string,
	files []source.File,
) (lifted map[string][]string, direct map[string][]string) {
	lifted = map[string][]string{}
	direct = map[string][]string{}
	for _, file := range files {
		addEdges(
			modulePath,
			file,
			lifted,
			direct,
		)
	}

	return compacted(lifted),
		compacted(direct)
}

func addEdges(
	modulePath string,
	file source.File,
	lifted map[string][]string,
	direct map[string][]string,
) {
	importer := filepath.Dir(file.Rel)
	for _, spec := range file.Syntax.Imports {
		from, to := siblingEdge(
			modulePath,
			importer,
			spec,
		)
		lifted[from] = append(
			lifted[from],
			to...,
		)
		direct[importer] = append(
			direct[importer],
			directTarget(
				importer,
				from,
				to,
			)...,
		)
	}
}

func directTarget(
	importer string,
	from string,
	to []string,
) (target []string) {
	if from == importer {
		return to
	}

	return nil
}

func siblingEdge(
	modulePath string,
	importer string,
	spec *ast.ImportSpec,
) (from string, to []string) {
	imported, inModule := importedDirectory(
		modulePath,
		spec,
	)
	if !inModule {
		return "",
			nil
	}

	from, isEdge := liftedImporter(
		importer,
		imported,
	)
	if !isEdge {
		return "",
			nil
	}

	return from,
		[]string{imported}
}

func liftedImporter(
	importer string,
	imported string,
) (from string, isEdge bool) {
	parent := filepath.Dir(imported)
	below, isBelowParent := relativeBelow(
		importer,
		parent,
	)
	isSiblingTree := imported != moduleRoot && isBelowParent
	if !isSiblingTree {
		return "",
			false
	}

	from = filepath.Join(
		parent,
		strings.Split(
			below,
			string(filepath.Separator),
		)[0],
	)
	return from,
		from != imported
}

func relativeBelow(
	directory string,
	ancestor string,
) (below string, isBelow bool) {
	if ancestor == moduleRoot {
		return directory,
			directory != moduleRoot
	}

	return strings.CutPrefix(
		directory,
		ancestor+string(filepath.Separator),
	)
}

func chainFindings(
	modulePath string,
	lifted map[string][]string,
	file source.File,
) (findings []report.Finding) {
	importer := filepath.Dir(file.Rel)
	for _, spec := range file.Syntax.Imports {
		from, to := siblingEdge(
			modulePath,
			importer,
			spec,
		)
		findings = append(
			findings,
			chainFinding(
				lifted,
				file,
				spec,
				from,
				to,
			)...,
		)
	}

	return findings
}

func chainFinding(
	lifted map[string][]string,
	file source.File,
	spec *ast.ImportSpec,
	from string,
	to []string,
) (findings []report.Finding) {
	isChain := len(to) > 0 && len(lifted[to[0]]) > 0
	if !isChain {
		return nil
	}

	sibling := filepath.Base(to[0])
	return []report.Finding{{
		Rule:     chainRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"imports sibling %s%s, which itself imports sibling %s: siblings form one layer; have the parent's container.go pass %s in",
			sibling,
			through(
				filepath.Dir(file.Rel),
				from,
			),
			filepath.Base(lifted[to[0]][0]),
			sibling,
		),
	}}
}

func through(
	importer string,
	from string,
) (note string) {
	if importer == from {
		return ""
	}

	return fmt.Sprintf(
		" (through %s)",
		filepath.Base(from),
	)
}

func wiringFindings(
	direct map[string][]string,
) (findings []report.Finding) {
	for _, directory := range slices.Sorted(maps.Keys(direct)) {
		findings = append(
			findings,
			wiringFinding(
				directory,
				direct[directory],
			)...,
		)
	}

	return findings
}

func wiringFinding(
	directory string,
	siblings []string,
) (findings []report.Finding) {
	if len(siblings) < wiringThreshold {
		return nil
	}

	names := []string{}
	for _, sibling := range siblings {
		names = append(
			names,
			filepath.Base(sibling),
		)
	}

	return []report.Finding{{
		Rule:     wiringRule,
		Location: filepath.ToSlash(directory),
		Snippet: fmt.Sprintf(
			"imports siblings %s: wiring siblings together is the parent's job; move it into %s and inject what this package needs through an interface it declares",
			strings.Join(
				names,
				nameSeparator,
			),
			path.Join(
				owner(directory),
				containerFile,
			),
		),
	}}
}

func owner(
	directory string,
) (parent string) {
	parent = filepath.Dir(directory)
	if source.IsInternalDirectory(parent) {
		return filepath.ToSlash(filepath.Dir(parent))
	}

	return filepath.ToSlash(parent)
}

func compacted(
	graph map[string][]string,
) (unique map[string][]string) {
	unique = map[string][]string{}
	for directory, targets := range graph {
		unique[directory] = slices.Compact(slices.Sorted(slices.Values(targets)))
	}

	return unique
}

func withLogic(
	direct map[string][]string,
	logic map[string]bool,
) (kept map[string][]string) {
	kept = map[string][]string{}
	for directory, siblings := range direct {
		kept[directory] = logicSiblings(
			siblings,
			logic,
		)
	}

	return kept
}

func logicSiblings(
	siblings []string,
	logic map[string]bool,
) (kept []string) {
	for _, sibling := range siblings {
		kept = append(
			kept,
			withLogicOnly(
				sibling,
				logic,
			)...,
		)
	}

	return kept
}

func withLogicOnly(
	sibling string,
	logic map[string]bool,
) (kept []string) {
	if !logic[sibling] {
		return nil
	}

	return []string{sibling}
}

func logicDirectories(
	files []source.File,
) (logic map[string]bool) {
	logic = map[string]bool{}
	for _, file := range files {
		directory := filepath.Dir(file.Rel)
		logic[directory] = logic[directory] || slices.ContainsFunc(
			source.Nodes[*ast.FuncDecl](file.Syntax),
			isLogic,
		)
	}

	return logic
}

func isLogic(
	function *ast.FuncDecl,
) (hasLogic bool) {
	isErrorMethod := function.Recv != nil && slices.Contains(
		[]string{"Error", "Unwrap", "Is", "As", "String"},
		function.Name.Name,
	)
	return function.Body != nil && !isErrorMethod
}

func importedDirectory(
	modulePath string,
	spec *ast.ImportSpec,
) (directory string, inModule bool) {
	importPath := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	if importPath == modulePath {
		return moduleRoot,
			true
	}

	imported, inModule := strings.CutPrefix(
		importPath,
		modulePath+"/",
	)
	return filepath.FromSlash(imported),
		inModule
}

func isMock(
	file source.File,
) (mock bool) {
	return source.IsMock(file.Rel)
}
