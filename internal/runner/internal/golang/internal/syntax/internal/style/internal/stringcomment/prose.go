package stringcomment

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	sentenceWords = 6
	wordTrim      = ".,;:!?()\"'"
	markerWords   = `\b(TODO|FIXME|HACK|XXX)\b`
	escapeWords   = `(?i)\b(hacky|hack|kludge|workaround|stopgap|band-?aid|for now|fix later|remove once|remove when|until we)\b`
)

func stringText(
	value ast.Expr,
) (text string, isText bool) {
	switch typed := value.(type) {
	case *ast.BasicLit:
		return literalText(typed)
	case *ast.ParenExpr:
		return stringText(typed.X)
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
	left, isLeft := stringText(expression.X)
	right, isRight := stringText(expression.Y)
	return left + right,
		expression.Op == token.ADD && isLeft && isRight
}

func isEscapeProse(
	text string,
) (escape bool) {
	return regexp.MustCompile(markerWords).MatchString(text) ||
		regexp.MustCompile(escapeWords).MatchString(text)
}

func isSentence(
	text string,
) (sentence bool) {
	words := 0
	for _, field := range strings.Fields(text) {
		words += wordCount(field)
	}

	return words >= sentenceWords
}

func wordCount(
	field string,
) (count int) {
	trimmed := strings.Trim(
		field,
		wordTrim,
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
