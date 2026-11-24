package report

import (
	"encoding/json"
	"fmt"
)

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
