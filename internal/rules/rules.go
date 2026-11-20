package rules

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

// Rule is one entry in the universal manifest: a single mechanically
// checkable rule, its enforcement mechanism, and the explanation an agent
// receives when it's violated.
type Rule struct {
	ID          string `yaml:"id"`
	Level       string `yaml:"level"`
	Text        string `yaml:"text"`
	Enforcement string `yaml:"enforcement"`
	Rationale   string `yaml:"rationale"`
	FixStrategy string `yaml:"fix_strategy"`
}

//go:embed rules.yaml
var registryYAML []byte

// Registry returns the full rule manifest, keyed by rule ID. It's parsed
// fresh from the embedded rules.yaml on every call rather than cached in a
// package-level variable, so the registry stays immutable package state
// instead of mutable global state.
func Registry() map[string]Rule {
	var reg map[string]Rule
	if err := yaml.Unmarshal(registryYAML, &reg); err != nil {
		panic("internal/rules: embedded rules.yaml failed to parse: " + err.Error())
	}
	return reg
}
