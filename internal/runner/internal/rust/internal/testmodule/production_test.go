package testmodule_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/clippy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/testmodule"
)

func TestProduction(
	t *testing.T,
) {
	root := t.TempDir()
	err := os.WriteFile(
		filepath.Join(
			root,
			"lib.rs",
		),
		[]byte("fn real() {\n    panic!(\"a\");\n}\n\n#[cfg(test)]\nmod tests {\n    fn fake() {\n        panic!(\"b\");\n    }\n}\n"),
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}

	findings, err := testmodule.Production(
		root,
		[]clippy.Warning{
			{Path: "lib.rs", Line: 2, Rule: "clippy::panic", Text: "panic"},
			{Path: "lib.rs", Line: 8, Rule: "clippy::panic", Text: "panic"},
			{Path: "lib.rs", Line: 2, Rule: "clippy::unwrap_used", Text: "unwrap"},
		},
	)
	locations := []string{}
	for _, finding := range findings {
		locations = append(
			locations,
			finding.Rule+" "+finding.Location,
		)
	}

	want := []string{"clippy::panic lib.rs:2", "clippy::unwrap_used lib.rs:2"}
	isMismatch := err != nil ||
		!slices.Equal(
			locations,
			want,
		)
	if isMismatch {
		t.Errorf(
			"Production() = %v, %v; want %v",
			locations,
			err,
			want,
		)
	}
}

func TestProductionUnreadableFile(
	t *testing.T,
) {
	_, err := testmodule.Production(
		t.TempDir(),
		[]clippy.Warning{{Path: "missing.rs", Line: 1}},
	)
	if err == nil {
		t.Error("Production() error = nil, want a read error")
	}
}
