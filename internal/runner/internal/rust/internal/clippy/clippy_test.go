package clippy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/clippy"
)

func TestClippy(
	t *testing.T,
) {
	cases := []struct {
		name    string
		script  string
		want    []string
		wantErr bool
	}{
		{
			name: "unwrap, expect, unwrap_err and panic warnings are reported",
			script: "printf '%s\\n' " +
				"'src/a.rs:3:5: warning: used `unwrap()` on a `Result` value' " +
				"'src/a.rs:4:5: warning: used `expect()` on an `Option` value' " +
				"'src/b.rs:7:1: warning: used `unwrap_err()` on a `Result` value' " +
				"'src/b.rs:9:1: warning: `panic` should not be present in production code'",
			want: []string{
				"src/a.rs:3",
				"src/a.rs:4",
				"src/b.rs:7",
				"src/b.rs:9",
			},
		},
		{
			name:   "unrelated warnings and test paths are skipped",
			script: "printf '%s\\n' 'src/a.rs:3:5: warning: this expression returns unit' 'tests/it.rs:2:1: warning: used `unwrap()` on a `Result` value'",
			want:   nil,
		},
		{
			name:    "a failed clippy run is an error",
			script:  "printf 'error: could not compile\\n'\nexit 101",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		fakeCargo(
			t,
			tc.script,
		)
		assertClippy(
			t,
			tc.name,
			tc.want,
			tc.wantErr,
		)
	}
}

func TestClippyWithoutCargo(
	t *testing.T,
) {
	t.Setenv(
		"PATH",
		t.TempDir(),
	)
	_, err := clippy.Clippy(t.TempDir())
	if err == nil {
		t.Error("Clippy() error = nil, want an error when cargo is not installed")
	}
}

func fakeCargo(
	t *testing.T,
	script string,
) {
	t.Helper()
	directory := t.TempDir()
	err := os.WriteFile(
		filepath.Join(
			directory,
			"cargo",
		),
		[]byte("#!/bin/sh\n"+script+"\n"),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"write fake cargo: %v",
			err,
		)
	}

	t.Setenv(
		"PATH",
		directory+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func assertClippy(
	t *testing.T,
	name string,
	want []string,
	wantErr bool,
) {
	t.Helper()
	warnings, err := clippy.Clippy(t.TempDir())
	if (err != nil) != wantErr {
		t.Errorf(
			"%s: error = %v, wantErr %v",
			name,
			err,
			wantErr,
		)
	}

	got := []string{}
	for _, warning := range warnings {
		got = append(
			got,
			fmt.Sprintf(
				"%s:%d",
				warning.Path,
				warning.Line,
			),
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: warnings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
