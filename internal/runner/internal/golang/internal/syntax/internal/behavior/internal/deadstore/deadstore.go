package deadstore

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule              = "R26"
	unreadFormat      = "%s from %s is never read before return: check it with if %s != nil and return it wrapped with %%w"
	overwrittenFormat = "%s from %s is overwritten before anyone reads it: check it with if %s != nil right after the call and return it wrapped with %%w"
	loopFormat        = "%s from %s is overwritten on the next loop iteration before anyone reads it: check %s inside the loop or accumulate with errors.Join"
)

func Deadstore(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	for _, declaration := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				declaration.Type,
				declaration.Body,
			)...,
		)
	}

	for _, literal := range source.Nodes[*ast.FuncLit](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				literal.Type,
				literal.Body,
			)...,
		)
	}

	return findings
}

func functionFindings(
	file source.File,
	signature *ast.FuncType,
	body *ast.BlockStmt,
) (findings []report.Finding) {
	if body == nil ||
		hasJumps(body) {
		return nil
	}

	tracker := flowFor(excludedNames(body))
	tracker.pushScope()
	declareSignature(
		tracker,
		signature,
	)
	visitList(
		tracker,
		body.List,
	)
	tracker.exit()
	tracker.popScope()
	for _, position := range slices.Sorted(maps.Keys(tracker.stores)) {
		findings = appendFinding(
			findings,
			file,
			tracker,
			tracker.stores[position],
		)
	}

	return findings
}

func declareSignature(
	tracker *flow,
	signature *ast.FuncType,
) {
	for _, name := range fieldNames(signature.Params) {
		tracker.declare(name)
	}

	for _, name := range fieldNames(signature.Results) {
		tracker.declare(name)
		tracker.results = append(
			tracker.results,
			tracker.lookup(name.Name),
		)
	}
}

func fieldNames(
	fields *ast.FieldList,
) (names []*ast.Ident) {
	if fields == nil {
		return nil
	}

	for _, field := range fields.List {
		names = append(
			names,
			field.Names...,
		)
	}

	return names
}

func hasJumps(
	body *ast.BlockStmt,
) (jumps bool) {
	for _, branch := range source.Nodes[*ast.BranchStmt](body) {
		jumps = jumps || branch.Tok == token.GOTO
	}

	return jumps || len(source.Nodes[*ast.LabeledStmt](body)) > 0
}

func excludedNames(
	body *ast.BlockStmt,
) (excluded map[string]bool) {
	excluded = map[string]bool{}
	for _, literal := range source.Nodes[*ast.FuncLit](body) {
		markNames(
			excluded,
			source.Nodes[*ast.Ident](literal),
		)
	}

	for _, address := range source.Nodes[*ast.UnaryExpr](body) {
		markNames(
			excluded,
			addressedNames(address),
		)
	}

	return excluded
}

func markNames(
	excluded map[string]bool,
	identifiers []*ast.Ident,
) {
	for _, identifier := range identifiers {
		excluded[identifier.Name] = true
	}
}

func addressedNames(
	address *ast.UnaryExpr,
) (names []*ast.Ident) {
	identifier, isIdentifier := ast.Unparen(address.X).(*ast.Ident)
	if address.Op != token.AND ||
		!isIdentifier {
		return nil
	}

	return []*ast.Ident{identifier}
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	tracker *flow,
	stored store,
) (extended []report.Finding) {
	format := snippetFormat(
		tracker,
		stored,
	)
	if stored.call == "" ||
		format == "" {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(stored.target),
			Snippet: fmt.Sprintf(
				format,
				stored.target.Name,
				stored.call,
				stored.target.Name,
			),
		},
	)
}

func snippetFormat(
	tracker *flow,
	stored store,
) (format string) {
	position := stored.target.Pos()
	if !tracker.used[position] {
		return unusedFormat(
			tracker,
			position,
		)
	}

	isLost := tracker.cycled[position] && !tracker.readWithin(
		stored.variable,
		stored.loop,
	)
	if !isLost {
		return ""
	}

	return loopFormat
}

func unusedFormat(
	tracker *flow,
	position token.Pos,
) (format string) {
	if tracker.killed[position] &&
		!tracker.exited[position] {
		return overwrittenFormat
	}

	return unreadFormat
}
