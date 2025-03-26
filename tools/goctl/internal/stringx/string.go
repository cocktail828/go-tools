package stringx

import (
	"bytes"
	"slices"
	"strings"
	"unicode"
)

// Title returns a string value with s[0] which has been convert into upper case that
// there are not empty input text
func Title(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Untitle returns a string value with s[0] which has been convert into lower case that
// there are not empty input text
func Untitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
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
		target = append(target, strings.ToLower(item))
	}
	return strings.Join(target, "_")
}

// IsEmptyOrSpace returns true if the length of the string value is 0 after call strings.TrimSpace, or else returns false
func IsEmptyOrSpace(s string) bool {
	if len(s) == 0 {
		return true
	}
	return strings.TrimSpace(s) == ""
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

func ContainsWhiteSpace(s string) bool {
	return slices.ContainsFunc([]rune(s), func(e rune) bool {
		return slices.Contains([]rune{'\n', '\t', '\f', '\v', ' '}, e)
	})
}

func TrimWhiteSpace(s string) string {
	return strings.NewReplacer(" ", "", "\t", "", "\n", "", "\f", "", "\r", "").Replace(s)
}

func IsEmptyStringOrWhiteSpace(s string) bool {
	return len(TrimWhiteSpace(s)) == 0
}
