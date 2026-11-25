package shape

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/concretetype"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/deadvalue"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/filelinelimit"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/methodfile"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/nostutter"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/onetypeperfile"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/packagehygiene"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/roleinterface"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/stateless"
)

func Shape(
	files []source.File,
	everyFile []source.File,
) (findings []report.Finding) {
	return slices.Concat(
		onetypeperfile.Onetypeperfile(source.HandWritten(files)),
		methodfile.Methodfile(source.HandWritten(files)),
		filelinelimit.Filelinelimit(everyFile),
		packagehygiene.Packagehygiene(source.HandWritten(files)),
		nostutter.Nostutter(source.HandWritten(files)),
		concretetype.Concretetype(files),
		stateless.Stateless(source.HandWritten(files)),
		roleinterface.Roleinterface(everyFile),
		deadvalue.Deadvalue(
			source.HandWritten(files),
			slices.Concat(
				files,
				everyFile,
			),
		),
	)
}
