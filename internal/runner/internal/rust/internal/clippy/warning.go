package clippy

// Warning is one panic-risk warning cargo clippy reported: the file path
// relative to the crate root, its 1-based line, the Agent Protection rule
// it breaks, and clippy's own message text.
type Warning struct {
	Path string
	Line int
	Rule string
	Text string
}
