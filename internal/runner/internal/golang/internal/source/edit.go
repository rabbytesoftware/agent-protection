package source

import (
	"cmp"
	"slices"
)

// Edit replaces the bytes of a file between the offsets Start and End with
// Text. An insertion has Start equal to End.
type Edit struct {
	Start int
	End   int
	Text  string
}

// Splice applies edits to content in offset order, insertions at one offset
// in the order given. An edit overlapping one already applied is dropped:
// the bytes it was computed against are gone, so the caller recomputes it
// on the result.
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
