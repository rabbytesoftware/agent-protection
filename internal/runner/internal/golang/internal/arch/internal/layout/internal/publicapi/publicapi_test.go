package publicapi_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/publicapi"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const queueStruct = "package jobq\n\ntype Queue struct{}\n\nfunc (q *Queue) Enqueue() {}\n"

type fixture struct {
	name  string
	files map[string]string
	want  []string
}

func TestPublicapi(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "method in the root package is reported",
				files: map[string]string{"queue.go": queueStruct},
				want:  []string{"R14 queue.go:5"},
			},
			{
				name:  "package main passes",
				files: map[string]string{"cmd/app/main.go": "package main\n\nfunc main() {\n\trun()\n}\n\nfunc run() {}\n"},
				want:  nil,
			},
			{
				name:  "internal package without internal children passes",
				files: map[string]string{"internal/work/work.go": "package work\n\nfunc helper() {}\n\ntype impl struct{}\n"},
				want:  nil,
			},
			{
				name: "module directory under internal is held to the API shape",
				files: map[string]string{
					"internal/svc/svc.go":            "package svc\n\nfunc helper() {}\n",
					"internal/svc/internal/a/a.go":   "package a\n\nfunc helper() {}\n",
					"internal/svc/internal/a/b/b.go": "package b\n\nfunc helper() {}\n",
				},
				want: []string{"R14 internal/svc/svc.go:3"},
			},
			{
				name: "mock files and mock directories pass",
				files: map[string]string{
					"store/mock_store.go":  "package store\n\nfunc helper() {}\n",
					"store/mocks/store.go": "package mocks\n\nfunc helper() {}\n",
				},
				want: nil,
			},
			{
				name:  "example trees pass",
				files: map[string]string{"examples/basic/basic.go": "package basic\n\nfunc helper() {}\n"},
				want:  nil,
			},
		},
	)
}

func TestPublicapiSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"queue.go":                    queueStruct,
			"store/store.go":              "package store\n\nfunc helper() {}\n",
			"store/internal/disk/disk.go": "package disk\n",
		},
	)
	assertSnippet(
		t,
		findings,
		"queue.go:5",
		"method Queue.Enqueue in public package jobq: implementation lives under internal/; expose an interface here and construct the implementation in container.go",
	)
	assertSnippet(
		t,
		findings,
		"store/store.go:3",
		"func helper in module root store: move it under store/internal/",
	)
}

func TestPublicapiRejectsMissingModule(
	t *testing.T,
) {
	_, err := publicapi.Publicapi(
		t.TempDir(),
		nil,
	)
	if err == nil {
		t.Error("Publicapi() error = nil, want one without go.mod")
	}
}

func assertFixtures(
	t *testing.T,
	fixtures []fixture,
) {
	t.Helper()
	for _, each := range fixtures {
		assertFindings(
			t,
			each.name,
			checked(
				t,
				each.files,
			),
			each.want,
		)
	}
}

func checked(
	t *testing.T,
	files map[string]string,
) (findings []report.Finding) {
	t.Helper()
	root := module(
		t,
		files,
	)
	parsed, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"parse fixture module: %v",
			err,
		)
	}

	findings, err = publicapi.Publicapi(
		root,
		parsed,
	)
	if err != nil {
		t.Fatalf(
			"Publicapi() error = %v",
			err,
		)
	}

	return findings
}

func module(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	write(
		t,
		filepath.Join(
			root,
			"go.mod",
		),
		"module example.com/m\n\ngo 1.24\n",
	)
	for name, content := range files {
		write(
			t,
			filepath.Join(
				root,
				name,
			),
			content,
		)
	}

	return root
}

func write(
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

func assertFindings(
	t *testing.T,
	name string,
	findings []report.Finding,
	want []string,
) {
	t.Helper()
	got := []string{}
	want = append(
		[]string{},
		want...,
	)
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func assertSnippet(
	t *testing.T,
	findings []report.Finding,
	location string,
	snippet string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Location+" "+finding.Snippet,
		)
	}

	if !slices.Contains(
		got,
		location+" "+snippet,
	) {
		t.Errorf(
			"no finding at %s with snippet %q in %v",
			location,
			snippet,
			got,
		)
	}
}
