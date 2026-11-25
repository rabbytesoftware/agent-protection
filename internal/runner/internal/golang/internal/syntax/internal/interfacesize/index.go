package interfacesize

import (
	"go/ast"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const pathSeparator = "/"

func declarations(
	files []source.File,
) (index map[string]declaration) {
	packageNames := declaredPackageNames(files)
	module := modulePath(files)
	index = map[string]declaration{}
	for _, file := range files {
		record(
			index,
			file,
			fileAliases(
				file,
				module,
				packageNames,
			),
		)
	}

	return index
}

func record(
	index map[string]declaration,
	file source.File,
	aliases map[string]string,
) {
	directory := packageDirectory(file)
	for _, spec := range interfaceSpecs(file.Syntax) {
		index[declarationKey(
			directory,
			spec.Name.Name,
		)] = declaration{
			directory: directory,
			spec:      spec,
			body:      interfaceBody(spec),
			aliases:   aliases,
		}
	}
}

func interfaceSpecs(
	syntax *ast.File,
) (specs []*ast.TypeSpec) {
	for _, decl := range syntax.Decls {
		specs = append(
			specs,
			declInterfaces(decl)...,
		)
	}

	return specs
}

func declInterfaces(
	decl ast.Decl,
) (specs []*ast.TypeSpec) {
	general, isGeneral := decl.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	for _, spec := range general.Specs {
		specs = appendInterface(
			specs,
			spec,
		)
	}

	return specs
}

func appendInterface(
	specs []*ast.TypeSpec,
	spec ast.Spec,
) (extended []*ast.TypeSpec) {
	typeSpec, isType := spec.(*ast.TypeSpec)
	if !isType {
		return specs
	}

	_, isInterface := typeSpec.Type.(*ast.InterfaceType)
	if !isInterface {
		return specs
	}

	return append(
		specs,
		typeSpec,
	)
}

func interfaceBody(
	spec *ast.TypeSpec,
) (body *ast.InterfaceType) {
	body, isInterface := spec.Type.(*ast.InterfaceType)
	if !isInterface {
		return nil
	}

	return body
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
