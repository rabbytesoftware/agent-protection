package deadstore

import (
	"go/token"
	"maps"
	"slices"
)

type reaching map[token.Pos][]token.Pos

func (r reaching) clone() (copied reaching) {
	if r == nil {
		return nil
	}

	copied = reaching{}
	for variable, stores := range r {
		copied[variable] = slices.Clone(stores)
	}

	return copied
}

func (r reaching) union(
	other reaching,
) (joined reaching) {
	if r == nil {
		return other.clone()
	}

	joined = r.clone()
	for variable, stores := range other {
		joined[variable] = slices.Compact(slices.Sorted(slices.Values(slices.Concat(
			joined[variable],
			stores,
		))))
	}

	return joined
}

func (r reaching) equal(
	other reaching,
) (same bool) {
	return (r == nil) == (other == nil) && maps.EqualFunc(
		r,
		other,
		slices.Equal[[]token.Pos],
	)
}

func joined(
	states []reaching,
) (merged reaching) {
	for _, state := range states {
		merged = merged.union(state)
	}

	return merged
}
