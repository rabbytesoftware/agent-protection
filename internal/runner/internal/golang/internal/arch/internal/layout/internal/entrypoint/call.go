package entrypoint

import (
	"go/ast"
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	importQuotes      = "\"`"
	exitCall          = "os.Exit"
	printCall         = "fmt.Fprintln"
	stderrName        = "Stderr"
	osQualifier       = "os"
	pathSeparator     = "/"
	domainDot         = "."
	blankName         = "_"
	configurationVerb = "reads configuration through "
	libraryVerb       = "computes through "
	unwiredVerb       = " without wiring its result"
	goPrefix          = "go-"
	goSuffix          = "-go"
)

func callVerb(
	context delegation,
	call *ast.CallExpr,
) (verb string) {
	switch callee := calleeOf(call.Fun).(type) {
	case *ast.Ident:
		return identifierCallVerb(
			context,
			callee,
		)
	case *ast.SelectorExpr:
		return selectorCallVerb(
			context,
			call,
			callee,
		)
	case *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType:
		return "converts"
	}

	return noLogic
}

func calleeOf(
	function ast.Expr,
) (callee ast.Expr) {
	switch typed := ast.Unparen(function).(type) {
	case *ast.IndexExpr:
		return ast.Unparen(typed.X)
	case *ast.IndexListExpr:
		return ast.Unparen(typed.X)
	}

	return ast.Unparen(function)
}

func identifierCallVerb(
	context delegation,
	callee *ast.Ident,
) (verb string) {
	isDeclared := slices.Contains(
		context.unit.names,
		callee.Name,
	) || slices.Contains(
		context.locals,
		callee.Name,
	)
	if isDeclared {
		return noLogic
	}

	return "calls " + callee.Name
}

func selectorCallVerb(
	context delegation,
	call *ast.CallExpr,
	callee *ast.SelectorExpr,
) (verb string) {
	imported, isImport := qualifierPath(
		context,
		callee.X,
	)
	if !isImport {
		return noLogic
	}

	qualified := packageNameOf(imported) + "." + callee.Sel.Name
	if !isStandard(
		context.unit.modulePath,
		imported,
	) {
		return foreignCallVerb(
			context,
			call,
			imported,
			callee.Sel.Name,
		)
	}

	if isAllowedCall(
		context.unit.role,
		imported+"."+callee.Sel.Name,
		call,
	) {
		return noLogic
	}

	return "calls " + qualified
}

func foreignCallVerb(
	context delegation,
	call *ast.CallExpr,
	imported string,
	name string,
) (verb string) {
	if isOwnModule(
		context.unit.modulePath,
		imported,
	) {
		return noLogic
	}

	target := imported + "." + name
	qualified := packageNameOf(imported) + "." + name
	if slices.ContainsFunc(
		configurationPackages(),
		isTargetOf(target),
	) {
		return configurationVerb + qualified
	}

	if slices.ContainsFunc(
		logicPackages(),
		isTargetOf(target),
	) {
		return libraryVerb + qualified
	}

	if slices.Contains(
		context.wiring,
		ast.Expr(call),
	) {
		return noLogic
	}

	return "calls " + qualified + unwiredVerb
}

func isTargetOf(
	target string,
) (matches func(string) bool) {
	return func(
		denied string,
	) (isDenied bool) {
		return target == denied || strings.HasPrefix(
			target,
			denied+pathSeparator,
		) || strings.HasPrefix(
			target,
			denied+domainDot,
		)
	}
}

func wiringCalls(
	body *ast.BlockStmt,
) (calls []ast.Expr) {
	for node := range ast.Preorder(body) {
		calls = append(
			calls,
			passedCalls(node)...,
		)
	}

	return calls
}

func passedCalls(
	node ast.Node,
) (calls []ast.Expr) {
	switch typed := node.(type) {
	case *ast.AssignStmt:
		return assignedCalls(typed)
	case *ast.ValueSpec:
		return onlyCalls(typed.Values)
	case *ast.ReturnStmt:
		return onlyCalls(typed.Results)
	case *ast.CallExpr:
		return onlyCalls(typed.Args)
	}

	return nil
}

func assignedCalls(
	statement *ast.AssignStmt,
) (calls []ast.Expr) {
	if !slices.ContainsFunc(
		statement.Lhs,
		isNamedTarget,
	) {
		return nil
	}

	return onlyCalls(statement.Rhs)
}

func isNamedTarget(
	target ast.Expr,
) (named bool) {
	identifier, isIdentifier := target.(*ast.Ident)
	return !isIdentifier || identifier.Name != blankName
}

func onlyCalls(
	expressions []ast.Expr,
) (calls []ast.Expr) {
	for _, expression := range expressions {
		calls = append(
			calls,
			callOf(expression)...,
		)
	}

	return calls
}

func callOf(
	expression ast.Expr,
) (calls []ast.Expr) {
	call, isCall := ast.Unparen(expression).(*ast.CallExpr)
	if !isCall {
		return nil
	}

	return []ast.Expr{call}
}

func isOwnModule(
	modulePath string,
	imported string,
) (own bool) {
	return imported == modulePath || strings.HasPrefix(
		imported,
		modulePath+pathSeparator,
	)
}

func qualifierPath(
	context delegation,
	expression ast.Expr,
) (imported string, isImport bool) {
	qualifier, isIdentifier := expression.(*ast.Ident)
	isQualifier := isIdentifier && !slices.Contains(
		context.locals,
		qualifier.Name,
	)
	if !isQualifier {
		return "",
			false
	}

	paths := importsNamed(
		context.unit.file.Syntax,
		qualifier.Name,
	)
	if len(paths) == 0 {
		return "",
			false
	}

	return paths[0],
		true
}

func isAllowedCall(
	fileRole role,
	qualified string,
	call *ast.CallExpr,
) (allowed bool) {
	if slices.Contains(
		allowedCalls(),
		qualified,
	) {
		return true
	}

	if fileRole != mainRole {
		return false
	}

	return qualified == exitCall || qualified == printCall && isStderr(call.Args)
}

func isStderr(
	arguments []ast.Expr,
) (matches bool) {
	if len(arguments) == 0 {
		return false
	}

	selector, isSelector := ast.Unparen(arguments[0]).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	qualifier, isIdentifier := selector.X.(*ast.Ident)
	return isIdentifier && qualifier.Name == osQualifier && selector.Sel.Name == stderrName
}

func isStandard(
	modulePath string,
	imported string,
) (standard bool) {
	first, _, _ := strings.Cut(
		imported,
		pathSeparator,
	)
	return !isOwnModule(
		modulePath,
		imported,
	) && !strings.Contains(
		first,
		domainDot,
	)
}

func importsNamed(
	syntax *ast.File,
	qualifier string,
) (paths []string) {
	for _, spec := range syntax.Imports {
		paths = append(
			paths,
			importNamed(
				spec,
				qualifier,
			)...,
		)
	}

	return paths
}

func importNamed(
	spec *ast.ImportSpec,
	qualifier string,
) (paths []string) {
	imported := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	name := packageNameOf(imported)
	if spec.Name != nil {
		name = spec.Name.Name
	}

	if name != qualifier {
		return nil
	}

	return []string{imported}
}

func allowedCalls() (qualified []string) {
	return []string{
		"errors.New",
		"errors.Join",
		"fmt.Errorf",
		"slices.Concat",
	}
}

func configurationPackages() (paths []string) {
	return []string{
		"github.com/spf13/pflag",
		"github.com/spf13/viper",
		"github.com/caarlos0/env",
		"github.com/kelseyhightower/envconfig",
		"github.com/sethvargo/go-envconfig",
		"github.com/joho/godotenv",
		"github.com/ilyakaznacheev/cleanenv",
		"github.com/knadh/koanf",
		"github.com/urfave/cli",
		"github.com/alecthomas/kong",
		"github.com/jessevdk/go-flags",
		"golang.org/x/sys/unix.Getenv",
		"golang.org/x/sys/unix.Environ",
		"golang.org/x/sys/windows.Getenv",
	}
}

func logicPackages() (paths []string) {
	return []string{
		"github.com/samber/lo",
		"github.com/samber/mo",
		"github.com/thoas/go-funk",
		"golang.org/x/exp/slices",
		"golang.org/x/exp/maps",
	}
}

func packageNameOf(
	imported string,
) (name string) {
	name = path.Base(imported)
	if majorVersion().MatchString(name) {
		name = path.Base(path.Dir(imported))
	}

	name = majorSuffix().ReplaceAllString(
		name,
		"",
	)
	name = strings.TrimPrefix(
		name,
		goPrefix,
	)
	return strings.TrimSuffix(
		name,
		goSuffix,
	)
}

func majorVersion() (pattern *regexp.Regexp) {
	return regexp.MustCompile(`^v[0-9]+$`)
}

func majorSuffix() (pattern *regexp.Regexp) {
	return regexp.MustCompile(`\.v[0-9]+$`)
}
