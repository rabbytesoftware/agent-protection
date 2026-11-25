package rust_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust"
)

const library = "fn real() {\n    None::<u8>.unwrap();\n}\n\n#[cfg(test)]\nmod tests {\n    fn fake() {\n        None::<u8>.unwrap();\n    }\n}\n"

func TestRust(
	t *testing.T,
) {
	root := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			root,
			"src/lib.rs",
		),
		library,
	)
	writeFile(
		t,
		filepath.Join(
			root,
			"src/main.rs",
		),
		"// note\nfn main() {}\n",
	)
	fakeCargo(
		t,
		"printf '%s\\n' 'src/lib.rs:2:5: warning: used `unwrap()` on an `Option` value' 'src/lib.rs:8:9: warning: used `unwrap()` on an `Option` value'",
	)
	findings, err := rust.Rust(
		root,
		os.Getenv,
	)
	if err != nil {
		t.Fatalf(
			"Rust() error = %v",
			err,
		)
	}

	got := locations(findings)
	if !slices.Equal(
		got,
		[]string{"src/lib.rs:2"},
	) {
		t.Errorf(
			"Rust() locations = %v, want only the production unwrap at src/lib.rs:2; comments are scanned by the runner",
			got,
		)
	}
}

func TestRustUnreadableWarnedFile(
	t *testing.T,
) {
	fakeCargo(
		t,
		"printf '%s\\n' 'src/gone.rs:2:5: warning: used `unwrap()` on an `Option` value'",
	)
	_, err := rust.Rust(
		t.TempDir(),
		os.Getenv,
	)
	if err == nil {
		t.Error("Rust() error = nil, want an error for a warned file that cannot be read")
	}
}

func TestRustClippyFailure(
	t *testing.T,
) {
	fakeCargo(
		t,
		"exit 101",
	)
	_, err := rust.Rust(
		t.TempDir(),
		os.Getenv,
	)
	if err == nil {
		t.Error("Rust() error = nil, want the clippy failure")
	}
}

func fakeCargo(
	t *testing.T,
	script string,
) {
	t.Helper()
	directory := t.TempDir()
	writeFile(
		t,
		filepath.Join(
			directory,
			"cargo",
		),
		"#!/bin/sh\n"+script+"\n",
	)
	err := os.Chmod(
		filepath.Join(
			directory,
			"cargo",
		),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"make fake cargo executable: %v",
			err,
		)
	}

	t.Setenv(
		"PATH",
		directory+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func writeFile(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"mkdir %s: %v",
			path,
			err,
		)
	}

	err = os.WriteFile(
		path,
		[]byte(content),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}

func locations(
	findings []report.Finding,
) (got []string) {
	for _, finding := range findings {
		got = append(
			got,
			finding.Location,
		)
	}

	return got
}
