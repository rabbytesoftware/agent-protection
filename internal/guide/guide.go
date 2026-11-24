package guide

import _ "embed"

//go:embed guide.txt
var text string

func Guide() (guide string) {
	return text
}
