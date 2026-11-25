package nopanic_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/nopanic"
)

func TestScope(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "Fatalf on a logger built with log.New is reported",
			code: "package p\n\nimport (\n\t\"log\"\n\t\"os\"\n)\n\nfunc f() {\n\tlogger := log.New(\n\t\tos.Stderr,\n\t\t\"\",\n\t\t0,\n\t)\n\tlogger.Fatalf(\"boom\")\n}\n",
			want: []string{"p.go:14"},
		},
		{
			name: "Panicln on log.Default() is reported",
			code: "package p\n\nimport \"log\"\n\nfunc f() {\n\tlog.Default().Panicln(\"boom\")\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "Fatal on a *log.Logger parameter is reported",
			code: "package p\n\nimport \"log\"\n\nfunc f(\n\tlogger *log.Logger,\n) {\n\tlogger.Fatal(\"boom\")\n}\n",
			want: []string{"p.go:8"},
		},
		{
			name: "Fatal on a var declared as a zap logger is reported",
			code: "package p\n\nimport \"go.uber.org/zap\"\n\nfunc f() {\n\tvar logger *zap.SugaredLogger\n\tlogger.Fatalf(\"boom\")\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "Panicf on a logger struct field is reported",
			code: "package p\n\nimport \"log\"\n\ntype service struct {\n\tlogger *log.Logger\n}\n\nfunc (s *service) run() {\n\ts.logger.Panicf(\"boom\")\n}\n",
			want: []string{"p.go:10"},
		},
		{
			name: "logger method taken as a value is reported",
			code: "package p\n\nimport \"log\"\n\nfunc f(\n\tlogger *log.Logger,\n) {\n\tstop := logger.Fatalln\n\tstop(\"boom\")\n}\n",
			want: []string{"p.go:8"},
		},
		{
			name: "Printf on a logger passes",
			code: "package p\n\nimport \"log\"\n\nfunc f(\n\tlogger *log.Logger,\n) {\n\tlogger.Printf(\"ok\")\n}\n",
			want: nil,
		},
		{
			name: "Fatal on a value that is not a logger passes",
			code: "package p\n\nfunc f(\n\tprobe checker,\n) {\n\tprobe.Fatal(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "Fatal on a non-logger field passes",
			code: "package p\n\ntype service struct {\n\tprobe checker\n}\n\nfunc (s *service) run() {\n\ts.probe.Fatal(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "Fatal on a value built by another package passes",
			code: "package p\n\nimport \"example.com/check\"\n\nfunc f() {\n\tprobe := check.New()\n\tprobe.Fatal(\"boom\")\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			nopanic.Nopanic([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
