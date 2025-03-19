package stringx

import (
	"bytes"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var WhiteSpace = []rune{'\n', '\t', '\f', '\v', ' '}

// IsEmptyOrSpace returns true if the length of the string value is 0 after call strings.TrimSpace, or else returns false
func IsEmptyOrSpace(s string) bool {
	if len(s) == 0 {
		return true
	}
	if strings.TrimSpace(s) == "" {
		return true
	}
	return false
}

// Lower calls the strings.ToLower
func Lower(s string) string {
	return strings.ToLower(s)
}

// Upper calls the strings.ToUpper
func Upper(s string) string {
	return strings.ToUpper(s)
}

// ReplaceAll calls the strings.ReplaceAll
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Title calls the cases.Title
func Title(s string) string {
	if IsEmptyOrSpace(s) {
		return s
	}
	return cases.Title(language.English, cases.NoLower).String(s)
}

// ToCamel converts the input text into camel case
func ToCamel(s string) string {
	list := splitBy(s, func(r rune) bool { return r == '_' }, true)

	var target []string
	for _, item := range list {
		target = append(target, Title(item))
	}
	return strings.Join(target, "")
}

// ToSnake converts the input text into snake case
func ToSnake(s string) string {
	list := splitBy(s, unicode.IsUpper, false)
	var target []string
	for _, item := range list {
		target = append(target, Lower(item))
	}
	return strings.Join(target, "_")
}

// Untitle return the original string if rune is not letter at index 0
func Untitle(s string) string {
	if IsEmptyOrSpace(s) {
		return s
	}
	r := rune(s[0])
	if !unicode.IsUpper(r) && !unicode.IsLower(r) {
		return s
	}
	return string(unicode.ToLower(r)) + s[1:]
}

// it will not ignore spaces
func splitBy(s string, fn func(r rune) bool, remove bool) []string {
	if IsEmptyOrSpace(s) {
		return nil
	}

	var list []string
	buffer := new(bytes.Buffer)
	for _, r := range s {
		if fn(r) {
			if buffer.Len() != 0 {
				list = append(list, buffer.String())
				buffer.Reset()
			}
			if !remove {
				buffer.WriteRune(r)
			}
			continue
		}
		buffer.WriteRune(r)
	}
	if buffer.Len() != 0 {
		list = append(list, buffer.String())
	}
	return list
}

func ContainsAny(s string, runes ...rune) bool {
	if len(runes) == 0 {
		return true
	}
	tmp := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		tmp[r] = struct{}{}
	}

	for _, r := range s {
		if _, ok := tmp[r]; ok {
			return true
		}
	}
	return false
}

func ContainsWhiteSpace(s string) bool {
	return ContainsAny(s, WhiteSpace...)
}

func IsWhiteSpace(text string) bool {
	if len(text) == 0 {
		return true
	}
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
