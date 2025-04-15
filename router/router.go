package router

import (
	"errors"
	"strings"
	"sync"

	"github.com/cocktail828/go-tools/z"
)

var (
	ErrNotFound   = errors.New("no such handler")
	ErrMissSchema = errors.New("missing protocol scheme")
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

type Context struct {
	Params
	Schema                string
	Location              string // Processed path
	URI                   string // Original URI
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
}

// Handler is a function that can be registered to a route to handle
// requests.
type Handler func(Context) error

// Router is a Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	mu   sync.RWMutex
	tree map[string]*Node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool
}

// New returns a new initialized Router.
// URI auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{
		tree:                  map[string]*Node{},
		RedirectTrailingSlash: true,
		RedirectFixedPath:     true,
	}
}

// Handler registers a new request handle with the given path and method.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(schema, path string, handle Handler) {
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	var root *Node
	z.WithLock(&r.mu, func() {
		var ok bool
		if root, ok = r.tree[schema]; !ok {
			root = &Node{}
			r.tree[schema] = root
		}
	})
	root.AddRoute(path, handle)
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(uri string) (Handler, Params, bool) {
	schema, path, found := strings.Cut(uri, ":/")
	if !found {
		return nil, nil, false
	}

	var root *Node
	var ok bool
	z.WithRLock(&r.mu, func() { root, ok = r.tree[schema] })
	if ok {
		return root.GetValue(path)
	}
	return nil, nil, false
}

func (r *Router) serve(root *Node, c Context) error {
	loc := c.Location
	if handle, ps, tsr := root.GetValue(loc); handle != nil {
		c.Params = ps
		return handle(c)
	} else if loc != "/" {
		if tsr && r.RedirectTrailingSlash {
			if len(loc) > 1 && loc[len(loc)-1] == '/' {
				loc = loc[:len(loc)-1]
			} else {
				loc = loc + "/"
			}
			c.RedirectTrailingSlash = true
			c.Location = loc
			return r.serve(root, c)
		}

		// Try to fix the request path
		if r.RedirectFixedPath {
			fixedPath, found := root.FindCaseInsensitivePath(cleanPath(loc), r.RedirectTrailingSlash)
			if found {
				c.RedirectFixedPath = true
				c.Location = string(fixedPath)
				return r.serve(root, c)
			}
		}
	}

	return ErrNotFound
}

// rpc://xxx/yyy -> rpc, /xxx/yyy
func (r *Router) ServeURI(uri string) error {
	if !strings.Contains(uri, "://") {
		return ErrMissSchema
	}

	schema, path, found := strings.Cut(uri, ":/")
	if !found {
		return ErrMissSchema
	}

	var root *Node
	var ok bool
	z.WithRLock(&r.mu, func() { root, ok = r.tree[schema] })
	if ok {
		return r.serve(root, Context{URI: uri, Location: path})
	}
	return ErrNotFound
}
