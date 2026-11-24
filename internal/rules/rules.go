package rules

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed rules.yaml
var manifest []byte

func Rules() (registry map[string]Rule, err error) {
	err = yaml.Unmarshal(
		manifest,
		&registry,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"parse embedded rules.yaml: %w",
				err,
			)
	}

	return registry,
		nil
}
