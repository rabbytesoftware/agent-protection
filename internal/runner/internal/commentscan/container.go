package commentscan

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/clex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/golex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/jslex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/linelex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lualex"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/markuplex"
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
			shlex.Justfile(),
			pylex.Graphql(),
			lualex.Lua(),
			clex.Proto(),
			clex.Css(),
			clex.Scss(),
			clex.Jsonc(),
			clex.Hcl(),
			markuplex.Markup(
				jslex.Jslex(),
				clex.Css(),
			),
			linelex.Ignore(),
			linelex.Ini(),
			linelex.Properties(),
		},
		linelex.Fallback(),
	)
}
