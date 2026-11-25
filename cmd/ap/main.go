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
	os.Exit(dispatch.Dispatch(
		os.Args,
		command.Check(
			runner.Runner,
			rulebook.Docs,
			guide.Guide,
		),
		command.Fix(runner.Fix),
		command.Markdown(os.ReadFile),
		command.Manual(guide.Guide),
	))
}
