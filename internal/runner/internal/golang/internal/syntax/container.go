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
		droppederror.Droppederror(everyFile),
		gofmtcheck.Gofmtcheck(everyFile),
		boolparam.Boolparam(files),
		cqs.Cqs(files),
		sideeffect.Sideeffect(files),
		packagehygiene.Packagehygiene(files),
		nostutter.Nostutter(files),
		onetypeperfile.Onetypeperfile(files),
		ctxfirstparam.Ctxfirstparam(files),
		filelinelimit.Filelinelimit(files),
		nopanic.Nopanic(files),
		deadstore.Deadstore(files),
		methodfile.Methodfile(files),
		interfacesize.Interfacesize(files),
		packagestate.Packagestate(files),
		stateless.Stateless(files),
		concretetype.Concretetype(files),
		roleinterface.Roleinterface(everyFile),
		realinfra.Realinfra(tests),
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
