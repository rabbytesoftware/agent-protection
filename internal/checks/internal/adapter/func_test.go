package adapter

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestFunc_CheckCallsTheWrappedFunction(t *testing.T) {
	called := false
	f := Func(func(root string) ([]report.Finding, error) {
		called = true
		return []report.Finding{{Rule: "X", Location: root}}, nil
	})

	findings, err := f.Check("some/root")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !called {
		t.Error("expected the wrapped function to be called")
	}
	if len(findings) != 1 || findings[0].Location != "some/root" {
		t.Errorf("expected the wrapped function's return value to pass through, got %v", findings)
	}
}
