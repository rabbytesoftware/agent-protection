package style

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/argumentperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/conditionperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/genericname"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/gofmtcheck"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/maxdepth"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/namedreturn"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/noelse"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/paragraph"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/paramperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/returnperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/stringcomment"
)

func Style(
	files []source.File,
) (findings []report.Finding) {
	return slices.Concat(
		paramperline.Paramperline(files),
		argumentperline.Argumentperline(files),
		conditionperline.Conditionperline(files),
		returnperline.Returnperline(files),
		paragraph.Paragraph(files),
		maxdepth.Maxdepth(files),
		commentpolicy.Commentpolicy(files),
		stringcomment.Stringcomment(files),
		noelse.Noelse(files),
		namedreturn.Namedreturn(files),
		genericname.Genericname(files),
		gofmtcheck.Gofmtcheck(files),
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
