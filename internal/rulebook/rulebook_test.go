package rulebook_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/rabbytesoftware/agent-protection/internal/rulebook"
)

func TestRules(
	t *testing.T,
) {
	registry, err := rulebook.Rules()
	if err != nil {
		t.Fatalf(
			"Rules() error = %v",
			err,
		)
	}

	for id, rule := range registry {
		assertComplete(
			t,
			id,
			rule,
		)
	}
}

func assertComplete(
	t *testing.T,
	id string,
	rule rulebook.Rule,
) {
	t.Helper()
	isIncomplete := rule.ID != id || rule.Level == "" || rule.Text == "" || rule.Rationale == "" || rule.FixStrategy == "" || rule.NotAFix == ""
	if isIncomplete {
		t.Errorf(
			"rule %q is incomplete or keyed under the wrong ID: %+v",
			id,
			rule,
		)
	}
}

func TestEveryEmittedRuleIsRegistered(
	t *testing.T,
) {
	registry, err := rulebook.Rules()
	if err != nil {
		t.Fatalf(
			"Rules() error = %v",
			err,
		)
	}

	pending := pendingRules(t)
	for id, rule := range pending {
		assertComplete(
			t,
			id,
			rule,
		)
		registry[id] = rule
	}

	scanned := &scanner{}
	err = filepath.WalkDir(
		filepath.Join(
			"..",
			"runner",
		),
		scanned.visit,
	)
	if err != nil {
		t.Fatalf(
			"walk runner sources: %v",
			err,
		)
	}

	for _, id := range scanned.ids {
		assertRegistered(
			t,
			registry,
			id,
		)
	}
}

type scanner struct {
	ids []string
}

func (s *scanner) visit(
	path string,
	entry fs.DirEntry,
	walkErr error,
) (err error) {
	isProduction := walkErr == nil &&
		!entry.IsDir() &&
		strings.HasSuffix(
			path,
			".go",
		) &&
		!strings.HasSuffix(
			path,
			"_test.go",
		)
	if !isProduction {
		return walkErr
	}

	content, err := os.ReadFile(path)
	s.ids = append(
		s.ids,
		matchedIDs(content)...,
	)
	return err
}

func matchedIDs(
	content []byte,
) (ids []string) {
	pattern := regexp.MustCompile(`"(R[0-9]{2})"`)
	for _, match := range pattern.FindAllSubmatch(
		content,
		-1,
	) {
		ids = append(
			ids,
			string(match[1]),
		)
	}

	return ids
}

func assertRegistered(
	t *testing.T,
	registry map[string]rulebook.Rule,
	id string,
) {
	t.Helper()
	_, isRegistered := registry[id]
	if !isRegistered {
		t.Errorf(
			"rule %s is emitted by the runner but missing from rules.yaml",
			id,
		)
	}
}

func pendingRules(
	t *testing.T,
) (pending map[string]rulebook.Rule) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(
		"..",
		"..",
		"RULES-*.yaml",
	))
	if err != nil {
		t.Fatalf(
			"glob pending rule files: %v",
			err,
		)
	}

	pending = map[string]rulebook.Rule{}
	for _, path := range paths {
		mergePending(
			t,
			pending,
			path,
		)
	}

	return pending
}

func mergePending(
	t *testing.T,
	pending map[string]rulebook.Rule,
	path string,
) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}

	err = yaml.Unmarshal(
		content,
		&pending,
	)
	if err != nil {
		t.Fatalf(
			"parse %s: %v",
			path,
			err,
		)
	}
}
