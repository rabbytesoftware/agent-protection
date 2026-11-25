package packagestate_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/packagestate"
)

func TestPackagestateEnvironment(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{name: "os.Getenv in an internal package is reported", rel: "internal/square/square.go", code: "package square\n\nimport \"os\"\n\nfunc factor() (value string) {\n\treturn os.Getenv(\"FACTOR\")\n}\n", want: []string{"R24 internal/square/square.go:6"}},
		{name: "os.LookupEnv bound as a value is reported", rel: "p.go", code: "package p\n\nimport \"os\"\n\nfunc lookup() (find func(string) (string, bool)) {\n\treturn os.LookupEnv\n}\n", want: []string{"R24 p.go:6"}},
		{name: "aliased unix.Getenv is reported", rel: "p.go", code: "package p\n\nimport sys \"golang.org/x/sys/unix\"\n\nfunc f() (value string, found bool) {\n\treturn sys.Getenv(\"X\")\n}\n", want: []string{"R24 p.go:6"}},
		{name: "global flag read is reported", rel: "p.go", code: "package p\n\nimport \"flag\"\n\nfunc f() (args []string) {\n\tflag.Parse()\n\treturn flag.Args()\n}\n", want: []string{"R24 p.go:6", "R24 p.go:7"}},
		{name: "viper global read is reported", rel: "p.go", code: "package p\n\nimport \"github.com/spf13/viper\"\n\nfunc f() (mode string) {\n\treturn viper.GetString(\"mode\")\n}\n", want: []string{"R24 p.go:6"}},
		{name: "versioned env loader is reported", rel: "p.go", code: "package p\n\nimport \"github.com/caarlos0/env/v11\"\n\nfunc f(\n\tcfg *config,\n) (err error) {\n\treturn env.Parse(cfg)\n}\n", want: []string{"R24 p.go:8"}},
		{name: "go-prefixed env loader is reported", rel: "p.go", code: "package p\n\nimport \"github.com/sethvargo/go-envconfig\"\n\nfunc f(\n\tcfg *config,\n) (err error) {\n\treturn envconfig.Process(nil, cfg)\n}\n", want: []string{"R24 p.go:8"}},
		{name: "env read in container.go is reported", rel: "container.go", code: "package arith\n\nimport \"os\"\n\nfunc Arith() (mode string) {\n\treturn os.ExpandEnv(\"$MODE\")\n}\n", want: []string{"R24 container.go:6"}},
		{name: "flag set parsing passed arguments passes", rel: "p.go", code: "package p\n\nimport \"flag\"\n\nfunc f(\n\targs []string,\n) (err error) {\n\tset := flag.NewFlagSet(\"ap\", flag.ContinueOnError)\n\treturn set.Parse(args)\n}\n", want: nil},
		{name: "viper instance construction passes", rel: "p.go", code: "package p\n\nimport \"github.com/spf13/viper\"\n\nfunc f() (v *viper.Viper) {\n\treturn viper.New()\n}\n", want: nil},
		{name: "other os calls pass", rel: "p.go", code: "package p\n\nimport \"os\"\n\nfunc f() (err error) {\n\treturn os.Remove(\"x\")\n}\n", want: nil},
		{name: "local variable named os passes", rel: "p.go", code: "package p\n\nimport \"os\"\n\nfunc f(\n\tos lookup,\n) (value string) {\n\treturn os.Getenv(\"X\")\n}\n", want: nil},
		{name: "env read in package main passes", rel: "main.go", code: "package main\n\nimport \"os\"\n\nfunc main() {\n\trun(os.Getenv)\n}\n", want: nil},
		{name: "env read in a mock passes", rel: "mocks/env.go", code: "package mocks\n\nimport \"os\"\n\nfunc f() (value string) {\n\treturn os.Getenv(\"X\")\n}\n", want: nil},
		{name: "env read in a test passes", rel: "p_test.go", code: "package p\n\nimport \"os\"\n\nfunc f() (value string) {\n\treturn os.Getenv(\"X\")\n}\n", want: nil},
		{name: "unknown third-party package passes", rel: "p.go", code: "package p\n\nimport \"example.com/vendor/v2\"\n\nfunc f() (value string) {\n\treturn vendor.Getenv(\"X\")\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagestate.Packagestate([]source.File{parsed(
				t,
				tc.rel,
				tc.code,
			)}),
			tc.want,
		)
	}
}
