package command

import "github.com/rabbytesoftware/agent-protection/internal/report"

type Tools struct {
	Run     func(root string) (findings []report.Finding, err error)
	Rewrite func(root string) (changed []string, err error)
	Docs    func() (docs map[string]report.RuleDoc, err error)
	Manual  func() (guide string)
}
