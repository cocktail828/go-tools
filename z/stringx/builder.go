package stringx

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// concurrency unsafe
type Array struct {
	lines []string
}

func (a Array) Join(sep string) string { return strings.Join(a.lines, sep) }
func (a Array) Uniq() Array            { return Array{lo.Uniq(a.lines)} }

func (a *Array) WriteString(s string) { a.lines = append(a.lines, s) }
func (a *Array) WriteStringf(format string, args ...any) {
	a.lines = append(a.lines, fmt.Sprintf(format, args...))
}

func (a *Array) Write(p []byte) (n int, err error) {
	a.lines = append(a.lines, string(p))
	return len(p), nil
}
