package registry

import "strings"

func CheckVersion(ver string) string {
	if ver == "" {
		return ""
	}

	ver = strings.ToLower(ver)
	if !strings.HasPrefix(ver, "v") {
		return "v" + ver
	}
	return ver
}
