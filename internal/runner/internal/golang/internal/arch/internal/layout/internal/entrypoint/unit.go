package entrypoint

import "github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"

type unit struct {
	file       source.File
	modulePath string
	directory  string
	role       role
	names      []string
	interfaces []string
}
