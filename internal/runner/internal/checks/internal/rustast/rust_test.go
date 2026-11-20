package rustast

import (
	"os"
	"testing"
)

func TestClassifyRustMessage(t *testing.T) {
	testCases := []struct {
		name    string
		msg     string
		want    string
		wantOK  bool
	}{
		{name: "unwrap on Result", msg: "used `unwrap()` on a `Result` value", want: "RUST-PANIC-RISK", wantOK: true},
		{name: "expect on Option", msg: "used `expect()` on an `Option` value", want: "RUST-PANIC-RISK", wantOK: true},
		{name: "bare panic in production code", msg: "`panic` should not be present in production code", want: "RUST-PANIC-RISK", wantOK: true},
		{name: "unrelated clippy message", msg: "this expression returns unit", want: "", wantOK: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := classifyRustMessage(tc.msg)
			if ok != tc.wantOK || (ok && got != tc.want) {
				t.Errorf("classifyRustMessage(%q) = (%q, %v), want (%q, %v)", tc.msg, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestFindInlineTestModuleRanges(t *testing.T) {
	dir := t.TempDir()
	src := "fn real() { panic!(\"a\"); }\n\n#[cfg(test)]\nmod tests {\n\tfn fake() { panic!(\"b\"); }\n}\n"
	path := dir + "/f.rs"
	writeFile(t, path, src)

	ranges := findInlineTestModuleRanges(path)
	if len(ranges) != 1 {
		t.Fatalf("expected exactly 1 test-module range, got %d: %v", len(ranges), ranges)
	}
	// line 1 (the real fn) must NOT be in range; line 5 (inside mod tests) must be.
	if inAnyRange(1, ranges) {
		t.Error("line 1 (real production code) should not be in the test-module range")
	}
	if !inAnyRange(5, ranges) {
		t.Error("line 5 (inside #[cfg(test)] mod tests) should be in the test-module range")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
