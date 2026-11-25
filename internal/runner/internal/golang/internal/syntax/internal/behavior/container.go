package behavior

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/boolparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/cqs"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/ctxfirstparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/deadstore"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/nopanic"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/packagestate"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/realinfra"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/sideeffect"
)

func Behavior(
	files []source.File,
	tests []source.File,
) (findings []report.Finding) {
	return slices.Concat(
		boolparam.Boolparam(source.HandWritten(files)),
		cqs.Cqs(files),
		sideeffect.Sideeffect(source.HandWritten(files)),
		ctxfirstparam.Ctxfirstparam(source.HandWritten(files)),
		nopanic.Nopanic(source.HandWritten(files)),
		deadstore.Deadstore(files),
		packagestate.Packagestate(files),
		droppederror.Droppederror(slices.Concat(
			files,
			tests,
		)),
		realinfra.Realinfra(source.HandWritten(tests)),
	)
}
