package component_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/component"
)

func TestCheck(
	t *testing.T,
) {
	cases := []struct {
		name    string
		markers []string
		want    []string
		wantErr bool
	}{
		{name: "the first present marker picks the checker and its owned extensions", markers: []string{component.RustMarker, component.GoMarker}, want: []string{"go", "comments:.go"}},
		{name: "a lone rust marker picks rust and scans every comment", markers: []string{component.RustMarker}, want: []string{"rust", "comments:"}},
		{name: "a failing checker is an error", markers: []string{component.TypescriptMarker}, wantErr: true},
		{name: "no marker is an error", markers: nil, wantErr: true},
	}
	for _, tc := range cases {
		findings, err := component.Check(
			rootWith(
				t,
				tc.markers,
			),
			os.Getenv,
			checkers(),
			scanned,
		)
		assertChecked(
			t,
			tc.name,
			findings,
			err,
			tc.want,
			tc.wantErr,
		)
	}
}

func TestCheckFailingCommentScan(
	t *testing.T,
) {
	_, err := component.Check(
		rootWith(
			t,
			[]string{component.GoMarker},
		),
		os.Getenv,
		checkers(),
		failingScan,
	)
	if err == nil {
		t.Error("Check() error = nil, want the comment scan failure")
	}
}

func TestFix(
	t *testing.T,
) {
	changed, err := component.Fix(
		rootWith(
			t,
			[]string{component.GoMarker},
		),
		fixed,
	)
	isWrong := err != nil ||
		!slices.Equal(
			changed,
			[]string{"fixed"},
		)
	if isWrong {
		t.Errorf(
			"Fix() on a Go root = %v, %v",
			changed,
			err,
		)
	}

	_, err = component.Fix(
		rootWith(
			t,
			[]string{component.RustMarker},
		),
		fixed,
	)
	if err == nil {
		t.Error("Fix() on a Rust root error = nil, want Go-only error")
	}
}

func checkers() (registered []component.Checker) {
	return []component.Checker{
		{Marker: component.GoMarker, Check: labelled("go"), Owned: []string{component.GoSource}},
		{Marker: component.RustMarker, Check: labelled("rust")},
		{Marker: component.TypescriptMarker, Check: failing},
	}
}

func labelled(
	label string,
) (check func(string, func(string) string) ([]report.Finding, error)) {
	return func(
		root string,
		environment func(string) string,
	) (findings []report.Finding, err error) {
		return []report.Finding{{Rule: label}},
			nil
	}
}

func failing(
	root string,
	environment func(string) string,
) (findings []report.Finding, err error) {
	return nil,
		errors.New("unreachable")
}

func scanned(
	root string,
	owned []string,
) (findings []report.Finding, err error) {
	return []report.Finding{{Rule: "comments:" + strings.Join(
			owned,
			",",
		)}},
		nil
}

func failingScan(
	root string,
	owned []string,
) (findings []report.Finding, err error) {
	return nil,
		errors.New("scan failed")
}

func rootWith(
	t *testing.T,
	markers []string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	for _, marker := range markers {
		writeMarker(
			t,
			filepath.Join(
				root,
				marker,
			),
		)
	}

	return root
}

func writeMarker(
	t *testing.T,
	path string,
) {
	t.Helper()
	err := os.WriteFile(
		path,
		nil,
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func fixed(
	root string,
) (changed []string, err error) {
	return []string{"fixed"},
		nil
}

func assertChecked(
	t *testing.T,
	name string,
	findings []report.Finding,
	err error,
	want []string,
	wantErr bool,
) {
	t.Helper()
	rules := []string{}
	for _, finding := range findings {
		rules = append(
			rules,
			finding.Rule,
		)
	}

	isMismatch := (err != nil) != wantErr ||
		len(want) > 0 && !slices.Equal(
			rules,
			want,
		)
	if isMismatch {
		t.Errorf(
			"%s: Check() = %v, %v; want %v, error %v",
			name,
			rules,
			err,
			want,
			wantErr,
		)
	}
}
