package packagehygiene

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	formatRule  = "R21"
	genericRule = "R22"
	underscore  = "_"
)

// Packagehygiene reports, at each file's package clause, a package name
// that is not lowercase or contains an underscore (R21), and a generic
// dumping-ground name such as util, common, helpers, base or misc (R22).
func Packagehygiene(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
		findings = append(
			findings,
			slices.Concat(
				formatFindings(file),
				genericFindings(file),
			)...,
		)
	}

	return findings
}

func formatFindings(
	file source.File,
) (findings []report.Finding) {
	name := file.Syntax.Name.Name
	isWellFormed := strings.ToLower(name) == name && !strings.Contains(
		name,
		underscore,
	)
	if isWellFormed {
		return nil
	}

	return []report.Finding{{
		Rule:     formatRule,
		Location: file.Location(file.Syntax.Name),
		Snippet: fmt.Sprintf(
			"package %q: rename it to a short lowercase noun without underscores",
			name,
		),
	}}
}

func genericFindings(
	file source.File,
) (findings []report.Finding) {
	name := file.Syntax.Name.Name
	if !slices.Contains(
		genericNames(),
		name,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     genericRule,
		Location: file.Location(file.Syntax.Name),
		Snippet: fmt.Sprintf(
			"package %q is a generic dumping ground: name it after what it does",
			name,
		),
	}}
}

func genericNames() (names []string) {
	return []string{"util", "utils", "common", "helpers", "helper", "base", "misc", "shared"}
}
