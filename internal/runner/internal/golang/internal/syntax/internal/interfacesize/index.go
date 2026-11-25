package interfacesize

import (
	"go/ast"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const pathSeparator = "/"

func catalogue(
	files []source.File,
) (indexed catalog) {
	packageNames := declaredPackageNames(files)
	module := modulePath(files)
	indexed = catalog{
		interfaces: map[string]declaration{},
		structures: map[string]structure{},
		methods:    map[string][]string{},
	}
	for _, file := range files {
		record(
			indexed,
			file,
			fileAliases(
				file,
				module,
				packageNames,
			),
		)
	}

	return indexed
}

func record(
	indexed catalog,
	file source.File,
	aliases map[string]string,
) {
	directory := packageDirectory(file)
	for _, spec := range typeSpecs(file.Syntax) {
		recordType(
			indexed,
			declaration{
				directory: directory,
				spec:      spec,
				aliases:   aliases,
			},
		)
	}

	for _, decl := range file.Syntax.Decls {
		recordMethod(
			indexed,
			directory,
			decl,
		)
	}
}

func recordType(
	indexed catalog,
	owner declaration,
) {
	key := declarationKey(
		owner.directory,
		owner.spec.Name.Name,
	)
	switch body := owner.spec.Type.(type) {
	case *ast.InterfaceType:
		owner.body = body
		indexed.interfaces[key] = owner
	case *ast.StructType:
		indexed.structures[key] = structure{
			owner: owner,
			body:  body,
		}
	}
}

func recordMethod(
	indexed catalog,
	directory string,
	decl ast.Decl,
) {
	function, isFunction := decl.(*ast.FuncDecl)
	isExposed := isFunction &&
		function.Recv != nil &&
		len(function.Recv.List) > 0 &&
		function.Name.IsExported()
	if !isExposed {
		return
	}

	receiver, isNamed := unindexed(unpointered(function.Recv.List[0].Type)).(*ast.Ident)
	if !isNamed {
		return
	}

	key := declarationKey(
		directory,
		receiver.Name,
	)
	indexed.methods[key] = append(
		indexed.methods[key],
		function.Name.Name,
	)
}

func typeSpecs(
	syntax *ast.File,
) (specs []*ast.TypeSpec) {
	for _, decl := range syntax.Decls {
		specs = append(
			specs,
			declTypes(decl)...,
		)
	}

	return specs
}

func declTypes(
	decl ast.Decl,
) (specs []*ast.TypeSpec) {
	general, isGeneral := decl.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	for _, spec := range general.Specs {
		specs = appendType(
			specs,
			spec,
		)
	}

	return specs
}

func appendType(
	specs []*ast.TypeSpec,
	spec ast.Spec,
) (extended []*ast.TypeSpec) {
	typeSpec, isType := spec.(*ast.TypeSpec)
	if !isType {
		return specs
	}

	return append(
		specs,
		typeSpec,
	)
}

func declaredPackageNames(
	files []source.File,
) (names map[string]string) {
	names = map[string]string{}
	for _, file := range files {
		names[packageDirectory(file)] = file.Syntax.Name.Name
	}

	return names
}

func declarationKey(
	directory string,
	name string,
) (key string) {
	return directory + pathSeparator + name
}

func packageDirectory(
	file source.File,
) (directory string) {
	return filepath.ToSlash(filepath.Dir(file.Rel))
}
