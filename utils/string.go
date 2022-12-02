package utils

import (
	"regexp"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

var stripAnsiEscapeRegexp = regexp.MustCompile(`(\x9B|\x1B\[)[0-?]*[ -/]*[@-~]`)

func stripAnsiEscape(s string) string {
	return stripAnsiEscapeRegexp.ReplaceAllString(s, "")
}

func RealLength(s string) int { //汉字长度为2
	return runewidth.StringWidth(stripAnsiEscape(s))
}

func Utf8Length(s string) int { //汉字长度为1
	return utf8.RuneCountInString(s)
}
