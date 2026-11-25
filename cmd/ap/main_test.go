package main

import (
	"os"
	"testing"
)

func TestMainExitsThroughDispatch(
	t *testing.T,
) {
	original := os.Args
	defer restoreArguments(original)
	defer assertExited(t)
	os.Args = []string{"ap", "guide"}
	main()
}

func restoreArguments(
	original []string,
) {
	os.Args = original
}

func assertExited(
	t *testing.T,
) {
	if recover() == nil {
		t.Error("main() returned without calling os.Exit")
	}
}
