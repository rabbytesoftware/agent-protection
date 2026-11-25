package entrypoint

type role string

const (
	containerRole role = "container"
	mainRole      role = "main"
	commandRole   role = "command"
	noRole        role = ""
)
