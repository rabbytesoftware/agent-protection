package report

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

const (
	markdownBudget  = 60000
	findingBullet   = "- "
	cleanVerdict    = "### Agent Protection: no violations\n"
	verdictFormat   = "### Agent Protection: %s across %s\n\n"
	truncatedFormat = "\n_Showing %d of %d findings. The full list is in the ap-report artifact of this run._\n"
	guidePointer    = "\nRun `ap guide` for the architecture these rules enforce. Fix by design: every rule's NOT A FIX is checked too.\n"
)

type Output struct {
	RuleDocs map[string]RuleDoc `json:"rule_docs"`
	Findings []Finding          `json:"findings"`
}

func (o Output) JSON() (document string, err error) {
	encoded, err := json.MarshalIndent(
		o,
		"",
		"  ",
	)
	if err != nil {
		return "",
			fmt.Errorf(
				"marshal report: %w",
				err,
			)
	}

	return string(encoded),
		nil
}

func (o Output) Text() (text string) {
	var builder strings.Builder
	for _, rule := range o.violatedRules() {
		builder.WriteString(o.section(rule))
	}

	fmt.Fprintf(
		&builder,
		"\n%d findings across %d unique rules\n",
		len(o.Findings),
		len(o.RuleDocs),
	)
	return builder.String()
}

func (o Output) violatedRules() (ids []string) {
	for _, finding := range o.Findings {
		ids = append(
			ids,
			finding.Rule,
		)
	}

	return slices.Compact(ids)
}

func (o Output) section(
	rule string,
) (section string) {
	doc := o.RuleDocs[rule]
	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"\n=== %s: %s ===\nWHY: %s\nFIX: %s\nNOT A FIX: %s\n\n",
		rule,
		doc.Text,
		doc.Rationale,
		doc.FixStrategy,
		doc.NotAFix,
	)
	for _, finding := range o.Findings {
		builder.WriteString(entry(
			finding,
			rule,
		))
	}

	return builder.String()
}

func entry(
	finding Finding,
	rule string,
) (line string) {
	if finding.Rule != rule {
		return ""
	}

	return fmt.Sprintf(
		"  x %s  %s\n",
		finding.Location,
		finding.Snippet,
	)
}

func (o Output) Markdown() (document string) {
	if len(o.Findings) == 0 {
		return cleanVerdict
	}

	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		verdictFormat,
		counted(
			len(o.Findings),
			"violation",
		),
		counted(
			len(o.violatedRules()),
			"rule",
		),
	)
	builder.WriteString("<details>\n<summary>Full report</summary>\n\n")
	builder.WriteString(o.breakdown())
	builder.WriteString(o.listing())
	builder.WriteString("\n</details>\n")
	builder.WriteString(guidePointer)
	return builder.String()
}

func (o Output) breakdown() (table string) {
	var builder strings.Builder
	builder.WriteString("| Rule | Violations | Rule text |\n|---|---:|---|\n")
	for _, rule := range o.violatedRules() {
		fmt.Fprintf(
			&builder,
			"| %s | %d | %s |\n",
			rule,
			o.countOf(rule),
			cell(o.RuleDocs[rule].Text),
		)
	}

	return builder.String()
}

func (o Output) listing() (listing string) {
	lines := []string{}
	for _, rule := range o.violatedRules() {
		lines = append(
			lines,
			o.ruleLines(rule)...,
		)
	}

	kept := fitting(
		lines,
		markdownBudget,
	)
	shown := bullets(kept)
	listing = strings.Join(
		kept,
		"",
	)
	if shown == len(o.Findings) {
		return listing
	}

	return listing + fmt.Sprintf(
		truncatedFormat,
		shown,
		len(o.Findings),
	)
}

func (o Output) ruleLines(
	rule string,
) (lines []string) {
	lines = append(
		lines,
		heading(
			rule,
			o.RuleDocs[rule],
		),
	)
	for _, finding := range o.Findings {
		lines = append(
			lines,
			bullet(
				finding,
				rule,
			)...,
		)
	}

	return lines
}

func (o Output) countOf(
	rule string,
) (count int) {
	for _, finding := range o.Findings {
		count += len(bullet(
			finding,
			rule,
		))
	}

	return count
}

func bullet(
	finding Finding,
	rule string,
) (lines []string) {
	if finding.Rule != rule {
		return nil
	}

	return []string{fmt.Sprintf(
		"%s`%s` `%s`\n",
		findingBullet,
		code(finding.Location),
		code(finding.Snippet),
	)}
}

func fitting(
	lines []string,
	budget int,
) (kept []string) {
	count := 0
	used := 0
	for count < len(lines) && used+len(lines[count]) <= budget {
		used += len(lines[count])
		count++
	}

	return lines[:count]
}

func bullets(
	lines []string,
) (count int) {
	return len(slices.DeleteFunc(
		slices.Clone(lines),
		isHeading,
	))
}

func isHeading(
	line string,
) (excluded bool) {
	return !strings.HasPrefix(
		line,
		findingBullet,
	)
}

func heading(
	rule string,
	doc RuleDoc,
) (line string) {
	if doc.Text == "" {
		return fmt.Sprintf(
			"\n#### %s\n\n",
			rule,
		)
	}

	return fmt.Sprintf(
		"\n#### %s: %s\n**Fix:** %s\n**Not a fix:** %s\n\n",
		rule,
		doc.Text,
		doc.FixStrategy,
		doc.NotAFix,
	)
}

func counted(
	count int,
	noun string,
) (phrase string) {
	if count == 1 {
		return fmt.Sprintf(
			"1 %s",
			noun,
		)
	}

	return fmt.Sprintf(
		"%d %ss",
		count,
		noun,
	)
}

func code(
	text string,
) (escaped string) {
	return strings.NewReplacer(
		"`",
		"'",
		"\n",
		" ",
	).Replace(text)
}

func cell(
	text string,
) (escaped string) {
	return strings.NewReplacer(
		"|",
		"\\|",
		"\n",
		" ",
	).Replace(text)
}
