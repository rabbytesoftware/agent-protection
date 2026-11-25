package stateless_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/stateless"
)

func TestStatelessAllowsInterfaceImplementations(
	t *testing.T,
) {
	clockInterface := "package clock\n\nimport \"time\"\n\ntype Clock interface {\n\tNow() (now time.Time)\n\tAfter(\n\t\tdelay time.Duration,\n\t) (fired <-chan time.Time)\n}\n"
	systemClock := "package system\n\nimport \"time\"\n\ntype systemClock struct{}\n\nfunc (systemClock) Now() (now time.Time) {\n\treturn time.Now()\n}\n\nfunc (systemClock) After(\n\tdelay time.Duration,\n) (fired <-chan time.Time) {\n\treturn time.After(delay)\n}\n"
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "empty clock implementing a module interface and returned as that interface passes",
			files: map[string]string{"clock/clock.go": clockInterface, "system/system_clock.go": systemClock, "system/system.go": "package system\n\ntype Clock interface {\n\tNow() (now time.Time)\n}\n\nfunc System() (clock Clock) {\n\treturn systemClock{}\n}\n"},
			want:  nil,
		},
		{
			name:  "empty clock whose constructor returns the struct is reported twice",
			files: map[string]string{"clock/clock.go": clockInterface, "system/system_clock.go": systemClock, "system/system.go": "package system\n\nfunc System() (clock *systemClock) {\n\treturn &systemClock{}\n}\n"},
			want:  []string{"R12 system/system.go:3", "R11 system/system_clock.go:5"},
		},
		{
			name:  "empty clock with a method no interface declares is reported",
			files: map[string]string{"clock/clock.go": clockInterface, "system/system_clock.go": systemClock, "system/tick.go": "package system\n\nfunc (systemClock) Tick() {\n}\n"},
			want:  []string{"R11 system/system_clock.go:5"},
		},
		{
			name:  "method name matching an interface with a different arity is reported",
			files: map[string]string{"clock/clock.go": "package clock\n\ntype Clock interface {\n\tNow(\n\t\tzone string,\n\t) (now int)\n}\n", "system/system_clock.go": "package system\n\ntype systemClock struct{}\n\nfunc (systemClock) Now() (now int) {\n\treturn 0\n}\n"},
			want:  []string{"R11 system/system_clock.go:3"},
		},
		{
			name:  "empty error and stringer types implementing stdlib interfaces pass",
			files: map[string]string{"p/missing.go": "package p\n\ntype missing struct{}\n\nfunc (missing) Error() (message string) {\n\treturn \"missing\"\n}\n\nfunc (missing) String() (text string) {\n\treturn \"missing\"\n}\n"},
			want:  nil,
		},
		{
			name:  "an interface declared only in a mock does not count",
			files: map[string]string{"p/mocks/runner.go": "package mocks\n\ntype Runner interface {\n\tRun()\n}\n", "p/service.go": "package p\n\ntype service struct{}\n\nfunc (service) Run() {\n}\n"},
			want:  []string{"R11 p/service.go:3"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stateless.Stateless(parsedFiles(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}
