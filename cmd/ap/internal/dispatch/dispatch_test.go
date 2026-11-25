package dispatch

import "testing"

const (
	checkCode    = 10
	fixCode      = 11
	markdownCode = 12
	guideCode    = 13
)

func TestDispatch(
	t *testing.T,
) {
	cases := []struct {
		arguments []string
		wantCode  int
	}{
		{arguments: []string{"ap"}, wantCode: usageExitCode},
		{arguments: []string{"ap", "guide"}, wantCode: guideCode},
		{arguments: []string{"ap", "guide", "extra"}, wantCode: usageExitCode},
		{arguments: []string{"ap", "check", "."}, wantCode: checkCode},
		{arguments: []string{"ap", "fix", "."}, wantCode: fixCode},
		{arguments: []string{"ap", "markdown", "ap-report.json"}, wantCode: markdownCode},
	}
	for _, tc := range cases {
		assertDispatch(
			t,
			tc.arguments,
			tc.wantCode,
		)
	}
}

func assertDispatch(
	t *testing.T,
	arguments []string,
	wantCode int,
) {
	t.Helper()
	code := Dispatch(
		arguments,
		fixedCommand(checkCode),
		fixedCommand(fixCode),
		fixedCommand(markdownCode),
		fixedCommand(guideCode),
	)
	if code != wantCode {
		t.Errorf(
			"Dispatch(%v) = %d, want %d",
			arguments,
			code,
			wantCode,
		)
	}
}

func TestCommandArguments(
	t *testing.T,
) {
	cases := []struct {
		arguments   []string
		wantCommand string
		wantTarget  string
		wantErr     bool
	}{
		{arguments: []string{"ap", "check", "."}, wantCommand: "check", wantTarget: "."},
		{arguments: []string{"ap", "markdown", "ap-report.json"}, wantCommand: "markdown", wantTarget: "ap-report.json"},
		{arguments: []string{"ap", "fix", "."}, wantCommand: "fix", wantTarget: "."},
		{arguments: []string{"ap", "check"}, wantErr: true},
		{arguments: []string{"ap", "lint", "."}, wantErr: true},
		{arguments: []string{"ap", "check", ".", "extra"}, wantErr: true},
	}
	for _, tc := range cases {
		assertCommandArguments(
			t,
			tc.arguments,
			tc.wantCommand,
			tc.wantTarget,
			tc.wantErr,
		)
	}
}

func assertCommandArguments(
	t *testing.T,
	arguments []string,
	wantCommand string,
	wantTarget string,
	wantErr bool,
) {
	t.Helper()
	command, target, err := commandArguments(arguments)
	isMismatch := command != wantCommand || target != wantTarget || (err != nil) != wantErr
	if isMismatch {
		t.Errorf(
			"commandArguments(%v) = %q, %q, %v; want %q, %q, error %v",
			arguments,
			command,
			target,
			err,
			wantCommand,
			wantTarget,
			wantErr,
		)
	}
}

type fixedCommand int

func (c fixedCommand) Run(
	target string,
) (code int) {
	return int(c)
}
