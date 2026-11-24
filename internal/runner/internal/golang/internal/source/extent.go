package source

import "go/token"

// Extent is where one item of a bracketed list starts and ends.
type Extent struct {
	Start token.Pos
	End   token.Pos
}
