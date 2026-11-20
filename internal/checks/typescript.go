package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

type eslintMessage struct {
	RuleID   string `json:"ruleId"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}

type eslintFileResult struct {
	FilePath string          `json:"filePath"`
	Messages []eslintMessage `json:"messages"`
}

// TSESLint runs eslint directly in root (root IS the package.json/eslint
// config directory -- one ap process, one language, one component folder).
// Surfaces every reported issue tagged by its real rule id
// (TS-ESLINT:<ruleId>) -- a pass-through of the project's existing rules,
// not yet a separately-authored Agent Protection TypeScript manifest.
func TSESLint(root string) []report.Finding {
	var findings []report.Finding
	cmd := exec.Command(bunBinary(), "x", "eslint", ".", "--format=json")
	cmd.Dir = root
	out, _ := cmd.CombinedOutput()
	var results []eslintFileResult
	if jerr := json.Unmarshal(extractJSON(out), &results); jerr != nil {
		return findings
	}
	for _, f := range results {
		for _, m := range f.Messages {
			if m.Severity == 0 || m.RuleID == "" {
				continue
			}
			rel := RelPath(root, f.FilePath)
			findings = append(findings, report.Finding{
				Rule:     "TS-ESLINT:" + m.RuleID,
				Location: fmt.Sprintf("%s:%d", rel, m.Line),
				Snippet:  m.Message,
			})
		}
	}
	return findings
}

func bunBinary() string {
	home, _ := os.UserHomeDir()
	candidate := filepath.Join(home, ".bun", "bin", "bun")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return "bun"
}

// extractJSON trims any non-JSON banner output some tools print before the
// actual JSON payload, by locating the first '[' or '{'.
func extractJSON(b []byte) []byte {
	s := string(b)
	for i, c := range s {
		if c == '[' || c == '{' {
			return []byte(s[i:])
		}
	}
	return b
}
