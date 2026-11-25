package publicapi_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/publicapi"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const storetestHelper = "package storetest\n\nfunc helper() {}\n"

func TestPublicapiTestSupportPackages(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name: "package imported only from other packages' tests passes",
				files: map[string]string{
					"store/storetest/storetest.go": storetestHelper,
					"store/memory_test.go":         "package store_test\n\nimport \"example.com/m/store/storetest\"\n",
				},
				want: nil,
			},
			{
				name: "package imported only by its own tests is reported",
				files: map[string]string{
					"store/storetest/storetest.go":      storetestHelper,
					"store/storetest/storetest_test.go": "package storetest_test\n\nimport \"example.com/m/store/storetest\"\n",
				},
				want: []string{"R14 store/storetest/storetest.go:3"},
			},
			{
				name: "package imported by tests and by production code is reported",
				files: map[string]string{
					"store/storetest/storetest.go": storetestHelper,
					"store/memory_test.go":         "package store_test\n\nimport \"example.com/m/store/storetest\"\n",
					"cmd/app/main.go":              "package main\n\nimport _ \"example.com/m/store/storetest\"\n\nfunc main() {}\n",
				},
				want: []string{"R14 store/storetest/storetest.go:3"},
			},
			{
				name: "root package imported only from tests passes",
				files: map[string]string{
					"helper.go":       "package m\n\nfunc helper() {}\n",
					"sub/sub_test.go": "package sub_test\n\nimport \"example.com/m\"\n",
				},
				want: nil,
			},
			{
				name: "package whose source imports testing passes",
				files: map[string]string{
					"store/storetest/storetest.go": "package storetest\n\nimport \"testing\"\n\nfunc run(t *testing.T) {}\n",
				},
				want: nil,
			},
			{
				name: "unparsable test files under testdata are ignored",
				files: map[string]string{
					"jobq.go":                 "package jobq\n\nfunc helper() {}\n",
					"testdata/broken_test.go": "package broken\n\nimport (\n",
				},
				want: []string{"R14 jobq.go:3"},
			},
		},
	)
}

func TestPublicapiRejectsUnparsableTestImports(
	t *testing.T,
) {
	root := module(
		t,
		map[string]string{"broken_test.go": "package broken\n\nimport (\n"},
	)
	files, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"parse fixture module: %v",
			err,
		)
	}

	_, err = publicapi.Publicapi(
		root,
		files,
	)
	if err == nil {
		t.Error("Publicapi() error = nil, want one for unparsable imports")
	}
}
