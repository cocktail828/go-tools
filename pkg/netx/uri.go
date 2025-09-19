package netx

import (
	"errors"
	"strings"
)

type RI struct {
	Schema string
	Path   string
	Query  string
}

func (r *RI) String() string {
	if r.Schema == "" {
		return r.Path + "?" + r.Query
	}
	return r.Schema + "://" + r.Path + "?" + r.Query
}

// ie.. mq://a/b?c=e
func ParseRI(uri string) (*RI, error) {
	ri := &RI{}
	schema, rest, found := strings.Cut(uri, "://")
	if found {
		ri.Schema = schema
	} else {
		rest = uri
	}

	path, query, found := strings.Cut(rest, "?")
	if !found {
		return nil, errors.New("invalid ri: missing path")
	}

	ri.Path = path
	if !strings.HasPrefix(path, "/") {
		ri.Path = "/" + path
	}
	ri.Query = query

	return ri, nil
}
