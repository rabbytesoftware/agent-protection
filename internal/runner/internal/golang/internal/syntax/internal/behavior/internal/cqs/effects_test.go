package cqs_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/cqs"
)

func TestCqsPackageEffects(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "mutation found through a chain of helpers is reported",
			code: "package p\n\ntype T struct {\n\tcalls []string\n}\n\nfunc (t *T) recordLocked(\n\tcall string,\n) {\n\tt.calls = append(t.calls, call)\n}\n\nfunc (t *T) record(\n\tcall string,\n) {\n\tt.recordLocked(call)\n}\n\nfunc (t *T) TicketFor() (n int) {\n\tt.record(\"TicketFor\")\n\treturn 0\n}\n",
			want: []string{"p.go:19"},
		},
		{
			name: "mutually recursive helpers that never write pass",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) even(\n\tn int,\n) (is bool) {\n\tif n == 0 {\n\t\treturn true\n\t}\n\n\treturn t.odd(n - 1)\n}\n\nfunc (t *T) odd(\n\tn int,\n) (is bool) {\n\tif n == 0 {\n\t\treturn false\n\t}\n\n\treturn t.even(n - 1)\n}\n",
			want: nil,
		},
		{
			name: "recursive helper that writes on its base case is reported with its callers",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) down(\n\tn int,\n) {\n\tif n == 0 {\n\t\tt.v = 1\n\t\treturn\n\t}\n\n\tt.down(n - 1)\n}\n\nfunc (t *T) Get() (v int) {\n\tt.down(3)\n\treturn t.v\n}\n",
			want: []string{"p.go:18"},
		},
		{
			name: "mutation through a helper called inside a closure is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc (t *T) Get() (v int) {\n\tdefer func() {\n\t\tt.bump()\n\t}()\n\treturn t.v\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "function declared without a body passes",
			code: "package p\n\nfunc external(\n\ttarget *int,\n) (n int)\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertCase(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func TestCqsAcrossFiles(
	t *testing.T,
) {
	helper := "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n"
	query := "package p\n\nfunc (t *T) Next() (v int) {\n\tt.bump()\n\treturn t.v\n}\n"
	cases := []struct {
		name       string
		helperPath string
		want       []string
	}{
		{
			name:       "helper declared in another file of the package is followed",
			helperPath: "a/helper.go",
			want:       []string{"a/query.go:3"},
		},
		{
			name:       "helper declared in another package directory is not followed",
			helperPath: "b/helper.go",
			want:       nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			cqs.Cqs([]source.File{
				parsedAt(
					t,
					tc.helperPath,
					helper,
				),
				parsedAt(
					t,
					"a/query.go",
					query,
				),
			}),
			tc.want,
		)
	}
}

func TestCqsQualifiedInterfaces(
	t *testing.T,
) {
	contract := "package ports\n\ntype Bus interface {\n\tPublish(\n\t\tevent string,\n\t)\n\tCount() (count int)\n}\n"
	query := "package a\n\nimport \"example.com/m/ports\"\n\ntype T struct {\n\tbus ports.Bus\n}\n\nfunc (t *T) Get() (v int) {\n\tt.bus.Publish(\"get\")\n\treturn t.bus.Count()\n}\n"
	assertLocations(
		t,
		"result-less method of an interface declared in another package is reported",
		cqs.Cqs([]source.File{
			parsedAt(
				t,
				"ports/bus.go",
				contract,
			),
			parsedAt(
				t,
				"a/query.go",
				query,
			),
		}),
		[]string{"a/query.go:9"},
	)
}
