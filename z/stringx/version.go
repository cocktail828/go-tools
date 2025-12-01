package stringx

import (
	"regexp"
	"strings"
)

var regexpVersion = regexp.MustCompile(`^(v?)((?:\d+\.){0,2}\d+)(\.(\d+))*(?:-rc\d+)?$`)

// return the canonical version without prefix "v"
func Version(in string) string {
	in = strings.ToLower(in)
	matches := regexpVersion.FindStringSubmatch(in)
	if len(matches) < 3 {
		return in
	}
	return strings.TrimPrefix(matches[2], "v")
}
