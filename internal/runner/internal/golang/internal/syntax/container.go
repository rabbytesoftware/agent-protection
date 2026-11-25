package syntax

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style"
)

func Syntax(
	files []source.File,
	tests []source.File,
) (findings []report.Finding) {
	everyFile := slices.Concat(
		source.HandWritten(files),
		source.HandWritten(tests),
	)
	return slices.Concat(
		style.Style(everyFile),
		shape.Shape(
			files,
			everyFile,
		),
		behavior.Behavior(
			files,
			tests,
		),
	)
}

func Edits(
	file source.File,
) (edits []source.Edit) {
	return style.Edits(file)
}
