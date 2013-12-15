package main

import (
	"html/template"
	"regexp"
	"strings"
	"unicode/utf8"
)

func unsafe(str string) template.HTML {
	return template.HTML(str)
}

var linebreakPattern, _ = regexp.Compile("\r?\n")

func linebreak(str string) string {
	// REVIEW: Should I use []byte instead of string?
	return string(linebreakPattern.ReplaceAll([]byte(str), []byte("<br>")))
}

func charCount(str string) int {
	withoutCr := strings.Replace(str, "\r\n", "\n", -1)
	return utf8.RuneCountInString(withoutCr)
}