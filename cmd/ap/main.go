package main

import (
	"os"

	"github.com/rabbytesoftware/agent-protection/cmd/ap/internal/command"
	"github.com/rabbytesoftware/agent-protection/cmd/ap/internal/dispatch"
	"github.com/rabbytesoftware/agent-protection/internal/guide"
	"github.com/rabbytesoftware/agent-protection/internal/rulebook"
	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

func main() {
	tools := command.Tools{
		Run:     runner.Runner,
		Rewrite: runner.Fix,
		Docs:    rulebook.Docs,
		Manual:  guide.Guide,
	}
	os.Exit(dispatch.Dispatch(
		os.Args,
		dispatch.Commands{
			Check:    tools.Check,
			Fix:      tools.Fix,
			Markdown: tools.Markdown,
			Guide:    tools.Guide,
		},
	))
}
