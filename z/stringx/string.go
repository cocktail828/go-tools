package stringx

import (
	"bytes"
	"strings"
	"unicode"
)

// Title returns a string value with s[0] which has been convert into upper case that
// there are not empty input text
func Title(s string) string {
	ss := strings.Split(s, " ")
	arr := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			arr = append(arr, strings.ToUpper(s[:1])+s[1:])
		}
	}
	return strings.Join(arr, " ")
}

// Untitle returns a string value with s[0] which has been convert into lower case that
// there are not empty input text
func Untitle(s string) string {
	ss := strings.Split(s, " ")
	arr := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			arr = append(arr, strings.ToLower(s[:1])+s[1:])
		}
	}
	return strings.Join(arr, " ")
}

// ToCamel converts the input text into camel case
func ToCamel(s string) string {
	list := splitBy(s, func(r rune) bool { return r == '_' }, true)
	for i, s := range list {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "_")
		if s != "" {
			list[i] = Title(s)
		}
	}
	return strings.Join(list, "")
}

// ToSnake converts the input text into snake case
func ToSnake(s string) string {
	list := splitBy(s, unicode.IsUpper, false)
	for i, s := range list {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "_")
		if s != "" {
			list[i] = strings.ToLower(s)
		}
	}
	return strings.Join(list, "_")
}

// it will not ignore spaces
func splitBy(s string, fn func(r rune) bool, remove bool) []string {
	if strings.TrimSpace(s) == "" {
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

// TrimSpaceAll will remove all space inside the string
func TrimSpaceAll(s string) string {
	return strings.NewReplacer(" ", "", "\t", "", "\n", "", "\v", "", "\f", "", "\r", "").Replace(s)
}

// HasPrefix reports whether the string s begins with prefix, but ignore the case.
func HasPrefix(s string, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// HasSuffix reports whether the string s ends with suffix, but ignore the case.
func HasSuffix(s string, suffix string) bool {
	return len(s) >= len(suffix) && strings.EqualFold(s[len(s)-len(suffix):], suffix)
}

// TrimPrefix returns s without the provided leading prefix string(ignore case).
// If s doesn't start with prefix, s is returned unchanged.
func TrimPrefix(s string, prefix string) string {
	if HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// TrimSuffix returns s without the provided trailing suffix string(ignore case).
// If s doesn't end with suffix, s is returned unchanged.
func TrimSuffix(s string, suffix string) string {
	if HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

func Trim(s string, cut string) string {
	return TrimPrefix(TrimSuffix(s, cut), cut)
}
