package workspace_test

import "testing"

func TestWorkspaceFindsDeeplyNestedModules(
	t *testing.T,
) {
	assertWorkspace(
		t,
		"modules nested inside nested modules are each reported",
		moduleWith(
			t,
			map[string]string{
				"a/go.mod":       "module a\n",
				"a/b/c/go.mod":   "module c\n",
				"a/b/c/x.go":     "package c\n",
				"pkg/inner.go":   "package pkg\n",
				"pkg/go.mod.bak": "module bak\n",
			},
		),
		[]string{"X a.go:1", "R70 a/b/c/go.mod", "R70 a/go.mod"},
	)
}
