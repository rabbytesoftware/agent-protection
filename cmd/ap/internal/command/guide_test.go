package command

import "testing"

func TestGuide(
	t *testing.T,
) {
	assertCode(
		t,
		"Guide()",
		realTools().Guide(),
		cleanExitCode,
	)
}
