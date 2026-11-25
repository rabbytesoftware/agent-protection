package rulebook_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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
