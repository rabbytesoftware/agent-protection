package syntax

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/argumentperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/boolparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/concretetype"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/conditionperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/cqs"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/ctxfirstparam"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/deadstore"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/filelinelimit"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/genericname"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/gofmtcheck"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/interfacesize"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/maxdepth"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/methodfile"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/namedreturn"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/noelse"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/nopanic"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/nostutter"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/onetypeperfile"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/packagehygiene"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/packagestate"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/paragraph"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/paramperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/realinfra"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/returnperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/roleinterface"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/sideeffect"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/stateless"
)

func Syntax(
	files []source.File,
	tests []source.File,
) (findings []report.Finding) {
	handWritten := source.HandWritten(files)
	everyFile := slices.Concat(
		handWritten,
		source.HandWritten(tests),
	)
	return slices.Concat(
		styleFindings(everyFile),
		droppederror.Droppederror(slices.Concat(
			files,
			tests,
		)),
		designFindings(
			files,
			handWritten,
		),
		roleinterface.Roleinterface(everyFile),
		realinfra.Realinfra(source.HandWritten(tests)),
	)
}

func styleFindings(
	everyFile []source.File,
) (findings []report.Finding) {
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
		gofmtcheck.Gofmtcheck(everyFile),
	)
}

func designFindings(
	files []source.File,
	handWritten []source.File,
) (findings []report.Finding) {
	return slices.Concat(
		boolparam.Boolparam(handWritten),
		cqs.Cqs(files),
		sideeffect.Sideeffect(handWritten),
		packagehygiene.Packagehygiene(handWritten),
		nostutter.Nostutter(handWritten),
		onetypeperfile.Onetypeperfile(handWritten),
		ctxfirstparam.Ctxfirstparam(handWritten),
		filelinelimit.Filelinelimit(handWritten),
		nopanic.Nopanic(handWritten),
		deadstore.Deadstore(files),
		methodfile.Methodfile(handWritten),
		interfacesize.Interfacesize(handWritten),
		packagestate.Packagestate(files),
		stateless.Stateless(handWritten),
		concretetype.Concretetype(files),
	)
}

func Edits(
	file source.File,
) (edits []source.Edit) {
	return slices.Concat(
		paramperline.Edits(file),
		argumentperline.Edits(file),
		paragraph.Edits(file),
	)
}
