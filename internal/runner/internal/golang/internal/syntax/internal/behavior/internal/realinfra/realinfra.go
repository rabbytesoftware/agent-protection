package realinfra

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R41"
	goPrefix      = "go-"
	snippetFormat = "unit test makes a live call to %s against a target it does not own: inject an in-memory fake through the interface the code depends on, or aim the call at an httptest.Server (srv.URL, srv.Client()) or a listener the test opens on 127.0.0.1:0 and reads back with Addr(); a localhost or host string literal is still live infrastructure"
)

func Realinfra(
	tests []source.File,
) (findings []report.Finding) {
	scopes := directoryScopes(tests)
	for _, file := range tests {
		findings = append(
			findings,
			fileFindings(
				file,
				scopes[filepath.Dir(file.Rel)],
			)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
	base scope,
) (findings []report.Finding) {
	imports := importedPaths(file.Syntax)
	for _, function := range nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				base,
				imports,
				function,
			)...,
		)
	}

	return findings
}

func functionFindings(
	file source.File,
	base scope,
	imports map[string]string,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	known := functionScope(
		base,
		imports,
		function,
	)
	for _, call := range nodes[*ast.CallExpr](function.Body) {
		findings = append(
			findings,
			callFindings(
				file,
				known,
				call,
			)...,
		)
	}

	return findings
}

func callFindings(
	file source.File,
	known scope,
	call *ast.CallExpr,
) (findings []report.Finding) {
	if !known.liveTarget(call) ||
		known.isExempt(call) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(call),
		Snippet: fmt.Sprintf(
			snippetFormat,
			types.ExprString(call.Fun),
		),
	}}
}
