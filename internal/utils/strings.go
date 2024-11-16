package utils

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"github.com/disgoorg/disgo/discord"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

type LocalizedString map[discord.Locale]string

func (str LocalizedString) String() string {
	for _, val := range str {
		return val
	}
	return "No localization found"
}

func (str LocalizedString) Using(locale discord.Locale) string {
	if val, ok := str[locale]; ok {
		return val
	}
	if val, ok := str[discord.LocaleEnglishUS]; ok {
		return val
	}
	return str.String()
}

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func AddSpaces(s string) string {
	buf := &bytes.Buffer{}
	for i, rune := range s {
		if unicode.IsUpper(rune) && i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteRune(rune)
	}
	return buf.String()
}
