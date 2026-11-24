package report

import (
	"encoding/json"
	"fmt"
)

// Decode parses an ap-report.json document back into an Output, so a report
// written by one run can be rendered in another form, such as the Markdown
// pull request comment, by a later step.
func Decode(
	document []byte,
) (output Output, err error) {
	err = json.Unmarshal(
		document,
		&output,
	)
	if err != nil {
		return Output{},
			fmt.Errorf(
				"decode report: %w",
				err,
			)
	}
	return output,
		nil
}
