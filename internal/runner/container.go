package runner

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/component"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript"
)

func Runner(
	root string,
) (findings []report.Finding, err error) {
	return component.Check(
		root,
		[]component.Checker{
			{Marker: component.GoMarker, Check: golang.Golang, Owned: []string{component.GoSource}},
			{Marker: component.RustMarker, Check: rust.Rust},
			{Marker: component.TypescriptMarker, Check: typescript.Typescript},
		},
		commentscan.Comments,
	)
}

func Fix(
	root string,
) (changed []string, err error) {
	return component.Fix(
		root,
		golang.Fix,
	)
}
