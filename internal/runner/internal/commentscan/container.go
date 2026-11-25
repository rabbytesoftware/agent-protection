package commentscan

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/golex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/jslex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/pylex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/rustlex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/scan"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/sqllex"
)

func Comments(
	root string,
	excluded []string,
) (findings []report.Finding, err error) {
	return scan.Scan(
		root,
		excluded,
		[]lexeme.Language{
			jslex.Jslex(),
			rustlex.Rustlex(),
			golex.Gosource(),
			golex.Gomod(),
			shlex.Shell(),
			shlex.Dockerfile(),
			shlex.Yaml(),
			shlex.Makefile(),
			pylex.Python(),
			pylex.Ruby(),
			pylex.Toml(),
			sqllex.Sqllex(),
		},
	)
}
