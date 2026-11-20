package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// apESLintConfig is Agent Protection's OWN TypeScript manifest, generated
// and run instead of the target repo's own eslint config -- same principle
// as GolangciLint's generated .yml for Go. Uses only ESLint core rules plus
// @typescript-eslint/parser (which any TS repo already has installed), so
// no custom plugin install is required of a consuming repo.
const apESLintConfig = `
import tsParser from '@typescript-eslint/parser';
export default [
  {
    files: ['**/*.{ts,tsx}'],
    ignores: ['**/node_modules/**', '**/dist/**', '**/build/**', '**/*.config.*', '**/*.gen.ts', '**/*.test.*', '**/*.spec.*', '**/e2e/**', '**/__tests__/**', '**/__mocks__/**'],
    languageOptions: {
      parser: tsParser,
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    rules: {
      'complexity': ['warn', 8],
      'max-depth': ['warn', 2],
      'max-lines-per-function': ['warn', { max: 40, skipBlankLines: true, skipComments: true }],
      'max-params': ['warn', 4],
      'no-magic-numbers': ['warn', { ignore: [0, 1, -1, 200, 201, 204, 400, 401, 403, 404, 409, 422, 429, 500, 502, 503], ignoreArrayIndexes: true, enforceConst: true }],
      'no-else-return': ['warn', { allowElseIf: false }],
      'no-lonely-if': 'warn',
      'no-console': 'warn',
      'no-param-reassign': ['warn', { props: true }],
      'no-var': 'warn',
      'prefer-const': 'warn',
      'eqeqeq': ['warn', 'always'],
      'max-classes-per-file': ['warn', 1],
      // TS-BOOL-PARAM removed: two esquery selector attempts either matched
      // return-type annotations and local variables (too broad) or matched
      // nothing at all including genuine hits (too narrow). A reliable
      // version needs real AST-walking (ts-morph/a custom rule), not a
      // guessed selector -- cut rather than ship broken or silent signal.
    },
  },
];
`

type eslintMessage2 struct {
	RuleID   string `json:"ruleId"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}

type eslintFileResult2 struct {
	FilePath string           `json:"filePath"`
	Messages []eslintMessage2 `json:"messages"`
}

var tsRuleIDToAP = map[string]string{
	"complexity":             "TS-PROXY-COMPLEXITY",
	"max-depth":              "TS-PROXY-COMPLEXITY",
	"max-lines-per-function": "TS-PROXY-COMPLEXITY",
	"max-params":             "TS-PROXY-COMPLEXITY",
	"no-magic-numbers":       "TS-PROXY-MAGIC-NUMBER",
	"no-else-return":         "TS-NO-ELSE",
	"no-lonely-if":           "TS-NO-ELSE",
	"no-console":             "TS-NO-DEBUG-OUTPUT",
	"no-param-reassign":      "TS-NO-PARAM-REASSIGN",
	"no-var":                 "TS-MODERN-SYNTAX",
	"prefer-const":           "TS-MODERN-SYNTAX",
	"eqeqeq":                 "TS-MODERN-SYNTAX",
	"max-classes-per-file":   "TS-ONE-TYPE-PER-FILE",
}

// TSManifest runs Agent Protection's own TypeScript rules (not the target
// repo's config) directly in root.
func TSManifest(root string) ([]report.Finding, error) {
	cfgPath := filepath.Join(root, ".ap-eslint-config.generated.mjs")
	if err := os.WriteFile(cfgPath, []byte(apESLintConfig), 0o644); err != nil {
		return nil, fmt.Errorf("write ts manifest config: %w", err)
	}
	defer os.Remove(cfgPath)

	cmd := exec.Command(bunBinary(), "x", "eslint", ".", "--config", cfgPath, "--format=json")
	cmd.Dir = root
	out, _ := cmd.CombinedOutput()

	var results []eslintFileResult2
	if jerr := json.Unmarshal(extractJSON(out), &results); jerr != nil {
		return nil, fmt.Errorf("ap TS manifest run produced no parseable output -- eslint/parser likely missing: %s", string(out))
	}

	var findings []report.Finding
	for _, f := range results {
		for _, m := range f.Messages {
			rule, known := tsRuleIDToAP[m.RuleID]
			if !known {
				continue
			}
			rel := RelPath(root, f.FilePath)
			findings = append(findings, report.Finding{
				Rule:     rule,
				Location: fmt.Sprintf("%s:%d", rel, m.Line),
				Snippet:  m.Message,
			})
		}
	}
	return findings, nil
}
