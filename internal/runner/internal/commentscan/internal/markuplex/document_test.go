package markuplex_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/markuplex"
)

func TestDocumentOffsetsEmbeddedSpans(
	t *testing.T,
) {
	text := []byte("<p/><script>a</script>")
	spans := markuplex.Markup(
		lexeme.Language{Spans: wholeBody},
		lexeme.Language{Spans: noSpans},
	).Spans(text)
	isWrong := len(spans) != 1 || string(text[spans[0].Start:spans[0].End]) != "a"
	if isWrong {
		t.Errorf(
			"Spans() = %v, want the script body only",
			spans,
		)
	}
}
