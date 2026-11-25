package deadvalue

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	proseWords    = 6
	excerptLength = 48
	ellipsis      = "..."
)

func proseOf(
	value ast.Expr,
) (excerpt string, isProse bool) {
	text, isText := stringValue(value)
	isProse = isText && (markerPattern().MatchString(text) || isSentence(text))
	if !isProse {
		return "",
			false
	}

	return shortened(text),
		true
}

func stringValue(
	value ast.Expr,
) (text string, isText bool) {
	switch typed := value.(type) {
	case *ast.BasicLit:
		return literalText(typed)
	case *ast.ParenExpr:
		return stringValue(typed.X)
	case *ast.BinaryExpr:
		return joinedText(typed)
	}

	return "",
		false
}

func literalText(
	literal *ast.BasicLit,
) (text string, isText bool) {
	if literal.Kind != token.STRING {
		return "",
			false
	}

	text, err := strconv.Unquote(literal.Value)
	return text,
		err == nil
}

func joinedText(
	expression *ast.BinaryExpr,
) (text string, isText bool) {
	left, isLeft := stringValue(expression.X)
	right, isRight := stringValue(expression.Y)
	return left + right,
		expression.Op == token.ADD && isLeft && isRight
}

func isSentence(
	text string,
) (sentence bool) {
	words := 0
	for _, field := range strings.Fields(text) {
		words += wordCount(field)
	}

	return words >= proseWords
}

func wordCount(
	field string,
) (count int) {
	trimmed := strings.TrimRight(
		field,
		".,;:!?",
	)
	isWord := trimmed != "" && !strings.ContainsFunc(
		trimmed,
		isNotWordRune,
	)
	if !isWord {
		return 0
	}

	return 1
}

func isNotWordRune(
	letter rune,
) (notWord bool) {
	return !unicode.IsLetter(letter) && letter != '\'' && letter != '-'
}

func shortened(
	text string,
) (excerpt string) {
	runes := []rune(text)
	if len(runes) <= excerptLength {
		return text
	}

	return string(runes[:excerptLength]) + ellipsis
}

func markerPattern() (pattern *regexp.Regexp) {
	return regexp.MustCompile(`\b(TODO|FIXME|HACK|XXX)\b`)
}
