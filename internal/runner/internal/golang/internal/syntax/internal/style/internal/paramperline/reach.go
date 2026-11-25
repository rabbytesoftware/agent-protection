package paramperline

import "go/token"

type reach struct {
	start token.Pos
	end   token.Pos
}

func (r reach) Pos() (position token.Pos) {
	return r.start
}

func (r reach) End() (position token.Pos) {
	return r.end
}
