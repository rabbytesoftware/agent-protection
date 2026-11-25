package component

import "github.com/rabbytesoftware/agent-protection/internal/report"

type Checker struct {
	Marker string
	Check  func(string) ([]report.Finding, error)
}
