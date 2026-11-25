package command

import "fmt"

func (t Tools) Guide() (code int) {
	fmt.Print(t.Manual())
	return cleanExitCode
}
