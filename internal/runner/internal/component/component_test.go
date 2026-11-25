package component_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
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
		{name: "the first present marker picks the checker", markers: []string{component.RustMarker, component.GoMarker}, want: []string{"go"}},
		{name: "a lone rust marker picks rust", markers: []string{component.RustMarker}, want: []string{"rust"}},
		{name: "no marker is an error", markers: nil, wantErr: true},
	}
	for _, tc := range cases {
		findings, err := component.Check(
			rootWith(
				t,
				tc.markers,
			),
			checkers(),
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
		{Marker: component.GoMarker, Check: labelled("go")},
		{Marker: component.RustMarker, Check: labelled("rust")},
		{Marker: component.TypescriptMarker, Check: failing},
	}
}

func labelled(
	label string,
) (check func(string) ([]report.Finding, error)) {
	return func(
		root string,
	) (findings []report.Finding, err error) {
		return []report.Finding{{Rule: label}},
			nil
	}
}

func failing(
	root string,
) (findings []report.Finding, err error) {
	return nil,
		errors.New("unreachable")
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
