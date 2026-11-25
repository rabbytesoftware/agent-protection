package packagestate

import (
	"fmt"
	"go/ast"
	"path"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	environmentFormat = "package %s reads configuration through %s: configuration arrives as parameters; take the value, or a lookup func, as a constructor parameter and let func main supply it"
	importQuotes      = "\"`"
	goPrefix          = "go-"
	goSuffix          = "-go"
	majorPrefix       = "v"
	digits            = "0123456789"
)

func environmentFindings(
	file source.File,
) (findings []report.Finding) {
	isExempt := file.Syntax.Name.Name == mainPackage || source.IsMock(file.Rel)
	if isExempt {
		return nil
	}

	imported := importedPaths(file.Syntax)
	for _, selector := range source.Nodes[*ast.SelectorExpr](file.Syntax) {
		findings = appendEnvironment(
			findings,
			file,
			selector,
			imported,
		)
	}

	return findings
}

func appendEnvironment(
	findings []report.Finding,
	file source.File,
	selector *ast.SelectorExpr,
	imported map[string]string,
) (appended []report.Finding) {
	qualifier, isIdent := selector.X.(*ast.Ident)
	isQualifier := isIdent && qualifier.Obj == nil
	if !isQualifier {
		return findings
	}

	importPath, isImport := imported[qualifier.Name]
	isRead := isImport && isConfigurationRead(
		importPath,
		selector.Sel.Name,
	)
	if !isRead {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(selector),
			Snippet: fmt.Sprintf(
				environmentFormat,
				file.Syntax.Name.Name,
				qualifier.Name+"."+selector.Sel.Name,
			),
		},
	)
}

func isConfigurationRead(
	importPath string,
	name string,
) (reads bool) {
	readers, isEnvironment := environmentReaders()[importPath]
	if isEnvironment {
		return slices.Contains(
			readers,
			name,
		)
	}

	unversioned := withoutMajor(importPath)
	instances, isGlobal := globalConfigurations()[unversioned]
	if isGlobal {
		return !slices.Contains(
			instances,
			name,
		)
	}

	return slices.Contains(
		environmentLoaders(),
		unversioned,
	)
}

func importedPaths(
	syntax *ast.File,
) (imported map[string]string) {
	imported = map[string]string{}
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(
			spec.Path.Value,
			importQuotes,
		)
		imported[importName(
			spec,
			importPath,
		)] = importPath
	}

	return imported
}

func importName(
	spec *ast.ImportSpec,
	importPath string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	name = path.Base(withoutMajor(importPath))
	return strings.TrimSuffix(
		strings.TrimPrefix(
			name,
			goPrefix,
		),
		goSuffix,
	)
}

func withoutMajor(
	importPath string,
) (unversioned string) {
	base := path.Base(importPath)
	version := strings.TrimPrefix(
		base,
		majorPrefix,
	)
	isVersioned := version != base && version != "" && strings.Trim(
		version,
		digits,
	) == ""
	if !isVersioned {
		return importPath
	}

	return path.Dir(importPath)
}

func environmentReaders() (readers map[string][]string) {
	return map[string][]string{
		"os": {
			"Getenv",
			"LookupEnv",
			"Environ",
			"ExpandEnv",
		},
		"syscall": {
			"Getenv",
			"Environ",
		},
		"golang.org/x/sys/unix": {
			"Getenv",
			"Environ",
		},
		"golang.org/x/sys/windows": {
			"Getenv",
			"Environ",
		},
	}
}

func globalConfigurations() (instances map[string][]string) {
	return map[string][]string{
		"flag":                       flagInstances(),
		"github.com/spf13/pflag":     flagInstances(),
		"github.com/spf13/viper":     viperInstances(),
		"github.com/namsral/flag":    flagInstances(),
		"github.com/peterbourgon/ff": nil,
	}
}

func flagInstances() (names []string) {
	return []string{
		"NewFlagSet",
		"FlagSet",
		"Flag",
		"Value",
		"Getter",
		"ErrorHandling",
		"ContinueOnError",
		"ExitOnError",
		"PanicOnError",
		"ErrHelp",
		"NormalizedName",
		"ParseErrorsWhitelist",
		"ParseErrorsAllowlist",
		"SliceValue",
		"UnquoteUsage",
	}
}

func viperInstances() (names []string) {
	return []string{
		"New",
		"NewWithOptions",
		"Viper",
		"Option",
		"KeyDelimiter",
		"EnvKeyReplacer",
		"DecoderConfigOption",
		"DecodeHook",
		"ConfigFileNotFoundError",
		"ConfigParseError",
		"UnsupportedConfigError",
		"RemoteConfigError",
		"StringReplacer",
	}
}

func environmentLoaders() (paths []string) {
	return []string{
		"github.com/caarlos0/env",
		"github.com/kelseyhightower/envconfig",
		"github.com/sethvargo/go-envconfig",
		"github.com/joho/godotenv",
		"github.com/ilyakaznacheev/cleanenv",
	}
}
