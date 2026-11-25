package component

import "github.com/rabbytesoftware/agent-protection/internal/report"

type Checker struct {
	Marker string
	Check  func(string, func(string) string) ([]report.Finding, error)
	Owned  []string
}
