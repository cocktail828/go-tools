package regular

import (
	"regexp"
	"strings"

	"github.com/cocktail828/go-tools/z"
)

type DirEntry interface {
	Name() string // base name of the file
}

type dirEntryImpl struct {
	name string
}

func (impl dirEntryImpl) Name() string { return impl.name }

type Filter func(DirEntry) bool

func WithPrefix(ext string) Filter {
	return func(de DirEntry) bool { return !strings.HasPrefix(de.Name(), ext) }
}

func WithSuffix(ext string) Filter {
	return func(de DirEntry) bool { return !strings.HasSuffix(de.Name(), ext) }
}

func WithRegular(fname ...string) Filter {
	return func(de DirEntry) bool { return !z.Contains(fname, de.Name()) }
}

func WithRegex(expr *regexp.Regexp) Filter {
	return func(de DirEntry) bool { return !expr.Match([]byte(de.Name())) }
}
