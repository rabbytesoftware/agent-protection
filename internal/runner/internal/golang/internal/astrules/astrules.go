package astrules

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/argumentperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/boolparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/conditionperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/cqs"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/ctxfirstparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/filelinelimit"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/genericname"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/maxdepth"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/namedreturn"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/noelse"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/noinfraintests"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/nopanic"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/nostutter"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/onetypeperfile"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/packagehygiene"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/paragraph"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/paramperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/returnperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/sideeffect"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/stateless"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

// Astrules runs every per-declaration syntax rule. Layout rules run on
// source and tests alike, design rules on source only, and the
// no-real-infrastructure rule on unit tests only.
func Astrules(
	files []source.File,
	tests []source.File,
) (findings []report.Finding) {
	everyFile := slices.Concat(
		files,
		tests,
	)
	return slices.Concat(
		paramperline.Paramperline(everyFile),
		argumentperline.Argumentperline(everyFile),
		conditionperline.Conditionperline(everyFile),
		returnperline.Returnperline(everyFile),
		paragraph.Paragraph(everyFile),
		maxdepth.Maxdepth(everyFile),
		commentpolicy.Commentpolicy(everyFile),
		noelse.Noelse(everyFile),
		namedreturn.Namedreturn(everyFile),
		genericname.Genericname(everyFile),
		boolparam.Boolparam(files),
		cqs.Cqs(files),
		sideeffect.Sideeffect(files),
		packagehygiene.Packagehygiene(files),
		nostutter.Nostutter(files),
		onetypeperfile.Onetypeperfile(files),
		ctxfirstparam.Ctxfirstparam(files),
		filelinelimit.Filelinelimit(files),
		nopanic.Nopanic(files),
		stateless.Stateless(files),
		noinfraintests.Noinfraintests(tests),
	)
}

// Edits returns the edits that bring file in line with the layout rules ap
// fix repairs: one parameter per line (R01), one argument per line (R52) and
// paragraph layout (R55). The edits only move line breaks, so they never
// change what the code means.
func Edits(
	file source.File,
) (edits []source.Edit) {
	return slices.Concat(
		paramperline.Edits(file),
		argumentperline.Edits(file),
		paragraph.Edits(file),
	)
}
