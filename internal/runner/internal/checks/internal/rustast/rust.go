package rustast

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

var rustLineRE = regexp.MustCompile(`^(.+?\.rs):(\d+):(\d+): warning: (.+)$`)

func rustLintToRule() map[string]string {
	return map[string]string{
		"unwrap": "RUST-PANIC-RISK",
		"expect": "RUST-PANIC-RISK",
		"panic":  "RUST-PANIC-RISK",
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func classifyRustMessage(msg string) (string, bool) {
	for frag, rule := range rustLintToRule() {
		if strings.Contains(msg, "`"+frag+"(") || strings.Contains(msg, "`"+frag+"_err(") ||
			strings.Contains(msg, frag+"` should not be present") {
			return rule, true
		}
	}
	return "", false
}

// RustClippy runs cargo clippy directly in root (root IS a Cargo.toml
// directory -- one ap process, one language, one component folder, the way
// this is actually wired in CI). Excludes separate _test.rs-style files AND
// inline `#[cfg(test)] mod tests { ... }` blocks -- the idiomatic Rust
// pattern of test doubles living in the same file as the code under test,
// confirmed as a real false-positive source (a Dead/Hangs test transport
// double inside a #[cfg(test)] module was otherwise flagged as if it were
// production code).
func RustClippy(root string) []report.Finding {
	var findings []report.Finding
	cmd := exec.Command("cargo", "clippy", "--all-targets", "--message-format=short", "--",
		"-W", "clippy::unwrap_used", "-W", "clippy::expect_used", "-W", "clippy::panic")
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, _ := cmd.CombinedOutput()

	testRangeCache := map[string][][2]int{}

	for _, line := range splitLines(string(out)) {
		m := rustLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		relPath := m[1]
		if strings.Contains(relPath, "test") {
			continue
		}
		lineNum, _ := strconv.Atoi(m[2])
		ranges, ok := testRangeCache[relPath]
		if !ok {
			ranges = findInlineTestModuleRanges(filepath.Join(root, relPath))
			testRangeCache[relPath] = ranges
		}
		if inAnyRange(lineNum, ranges) {
			continue
		}
		rule, ok2 := classifyRustMessage(m[4])
		if !ok2 {
			continue
		}
		findings = append(findings, report.Finding{
			Rule:     rule,
			Location: fmt.Sprintf("%s:%s", relPath, m[2]),
			Snippet:  m[4],
		})
	}
	return findings
}

// findInlineTestModuleRanges returns [startLine, endLine] pairs for every
// `#[cfg(test)]` attribute immediately followed by `mod ... {`, tracking
// brace depth from that mod's opening brace to its matching close.
func findInlineTestModuleRanges(path string) [][2]int {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	var ranges [][2]int
	for i := 0; i < len(lines); i++ {
		if r, ok := testModuleRangeStartingAt(lines, i); ok {
			ranges = append(ranges, r)
		}
	}
	return ranges
}

// testModuleRangeStartingAt reports the [startLine, endLine] range of the
// `#[cfg(test)] mod ... { ... }` block beginning at lines[i], if any.
func testModuleRangeStartingAt(lines []string, i int) ([2]int, bool) {
	if !strings.Contains(lines[i], "#[cfg(test)]") {
		return [2]int{}, false
	}
	j := skipBlankLines(lines, i+1)
	if j >= len(lines) || !strings.Contains(lines[j], "mod ") {
		return [2]int{}, false
	}
	start := j + 1
	end := findModuleCloseLine(lines, j)
	return [2]int{start, end}, true
}

// skipBlankLines returns the index of the first non-blank line at or after
// from.
func skipBlankLines(lines []string, from int) int {
	j := from
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	return j
}

// findModuleCloseLine tracks brace depth from openLine's opening brace and
// returns the 1-indexed line number where it closes (or the file's end if
// unterminated).
func findModuleCloseLine(lines []string, openLine int) int {
	depth := strings.Count(lines[openLine], "{") - strings.Count(lines[openLine], "}")
	for k := openLine + 1; k < len(lines) && depth > 0; k++ {
		depth += strings.Count(lines[k], "{") - strings.Count(lines[k], "}")
		if depth <= 0 {
			return k + 1
		}
	}
	return len(lines)
}

func inAnyRange(line int, ranges [][2]int) bool {
	for _, r := range ranges {
		if line >= r[0] && line <= r[1] {
			return true
		}
	}
	return false
}
