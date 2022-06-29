package main

import (
	"regexp"
)

var markdownSpecialCharacterMatcher = regexp.MustCompile(`[\\_\*\~]`)

func escapeString(s string) string {
	return markdownSpecialCharacterMatcher.ReplaceAllStringFunc(s, func(s string) string {
		return `\` + s
	})
}

func boldString(s string) string {
	return "**" + s + "**"
}

func italicString(s string) string {
	return "_" + s + "_"
}

func strikeString(s string) string {
	return "~~" + s + "~~"
}
