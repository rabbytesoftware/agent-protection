package source

import (
	"cmp"
	"slices"
)

type Edit struct {
	Start int
	End   int
	Text  string
}

func Splice(
	content []byte,
	edits []Edit,
) (spliced []byte) {
	cursor := 0
	for _, edit := range disjoint(edits) {
		spliced = slices.Concat(
			spliced,
			content[cursor:edit.Start],
			[]byte(edit.Text),
		)
		cursor = edit.End
	}

	return slices.Concat(
		spliced,
		content[cursor:],
	)
}

func disjoint(
	edits []Edit,
) (kept []Edit) {
	for _, edit := range slices.SortedStableFunc(
		slices.Values(edits),
		byPosition,
	) {
		kept = appendDisjoint(
			kept,
			edit,
		)
	}

	return kept
}

func appendDisjoint(
	kept []Edit,
	edit Edit,
) (extended []Edit) {
	isOverlapping := len(kept) > 0 && edit.Start < kept[len(kept)-1].End
	if isOverlapping {
		return kept
	}

	return append(
		kept,
		edit,
	)
}

func byPosition(
	first Edit,
	second Edit,
) (order int) {
	return cmp.Or(
		cmp.Compare(
			first.Start,
			second.Start,
		),
		cmp.Compare(
			first.End,
			second.End,
		),
	)
}
