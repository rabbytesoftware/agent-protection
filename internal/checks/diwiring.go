package checks

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

var trivialConventionMethod = map[string]bool{
	"TableName": true, "MarshalJSON": true, "UnmarshalJSON": true,
	"Scan": true, "Value": true, "String": true, "Error": true, "GoString": true,
}

var wiringFileHints = []string{"container.go", "wire.go", "main.go", "new.go"}

func isWiringFile(path string) bool {
	base := filepath.Base(path)
	for _, hint := range wiringFileHints {
		if base == hint || strings.HasSuffix(base, "_"+hint) || strings.Contains(base, "container") {
			return true
		}
	}
	return false
}

// DIWiring flags R37: a concrete struct literal for a package-local type
// constructed outside that type's own declaring file and outside a
// recognized wiring/composition file.
func DIWiring(root string, files []ParsedFile) []report.Finding {
	// group files by directory (package) since type names are only unique
	// within a package
	byDir := map[string][]ParsedFile{}
	for _, pf := range files {
		dir := filepath.Dir(pf.Path)
		byDir[dir] = append(byDir[dir], pf)
	}

	var findings []report.Finding
	for _, pkgFiles := range byDir {
		declFile := map[string]string{}  // type name -> declaring file path
		hasMethods := map[string]bool{}  // type name -> has at least one method
		for _, pf := range pkgFiles {
			for _, decl := range pf.AST.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						if _, isStruct := ts.Type.(*ast.StructType); isStruct {
							declFile[ts.Name.Name] = pf.Path
						}
					}
				case *ast.FuncDecl:
					if d.Recv == nil || trivialConventionMethod[d.Name.Name] {
						continue // TableName/MarshalJSON/Scan/Value/String/Error are
						// ORM and marshalling conventions attached to plain data
						// types; having only one of these isn't "behavior" in the
						// DI sense -- confirmed false positive against a real codebase's
						// GORM row types (arrowRow.TableName()).
					}
					for _, f := range d.Recv.List {
						switch t := f.Type.(type) {
						case *ast.StarExpr:
							if ident, ok := t.X.(*ast.Ident); ok {
								hasMethods[ident.Name] = true
							}
						case *ast.Ident:
							hasMethods[t.Name] = true
						}
					}
				}
			}
		}
		// Only flag types that have behavior (at least one method) -- a plain
		// data/DTO struct constructed where it's used is normal, not a
		// hidden-dependency smell. This is the fix for the false-positive
		// class confirmed against a real codebase's DTO package.
		for name := range declFile {
			if !hasMethods[name] {
				delete(declFile, name)
			}
		}
		for _, pf := range pkgFiles {
			if isWiringFile(pf.Path) {
				continue
			}
			ast.Inspect(pf.AST, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				ident, ok := lit.Type.(*ast.Ident)
				if !ok {
					return true
				}
				declaredIn, known := declFile[ident.Name]
				if !known || declaredIn == pf.Path {
					return true
				}
				pos := pf.FSet.Position(lit.Pos())
				findings = append(findings, report.Finding{
					Rule:     "R37",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
					Snippet:  fmt.Sprintf("%s{...} constructed outside its own file and outside a wiring file", ident.Name),
				})
				return true
			})
		}
	}
	return findings
}
