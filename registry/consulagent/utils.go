package consulagent

import (
	"strings"
)

func combine(args ...string) string {
	return strings.Join(args, "#")
}
