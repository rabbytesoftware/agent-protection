package command

import "fmt"

type manual struct {
	text func() (guide string)
}

func Manual(
	text func() (guide string),
) (command *manual) {
	return &manual{text: text}
}

func (m *manual) Run(
	target string,
) (code int) {
	fmt.Print(m.text())
	return cleanExitCode
}
