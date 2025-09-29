package dsn

import (
	"strings"
)

type RI struct {
	Scheme string
	Path   string
	Query  string
}

func (ri *RI) String() string {
	if ri.Scheme == "" {
		return ri.Path + "?" + ri.Query
	}
	return ri.Scheme + "://" + ri.Path + "?" + ri.Query
}

// ie.. mq://a/b?c=e
func Parse(uri string) (ri RI) {
	scheme, rest, found := strings.Cut(uri, "://")
	if found {
		ri.Scheme = scheme
	} else {
		rest = uri
	}

	ri.Path, ri.Query, _ = strings.Cut(rest, "?")
	if !strings.HasPrefix(ri.Path, "/") {
		ri.Path = "/" + ri.Path
	}
	return
}
