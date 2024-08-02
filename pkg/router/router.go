package router

import (
	"sync"
)

type Handler func(path string, ps Params)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// Get returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) Get(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// Router is a Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	mu   sync.RWMutex
	root *node
}

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{root: new(node)}
}

// Register registers a new request handle with the given path.
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Register(path string, handle Handler) {
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.root.addRoute(path, handle)
}

// Lookup allows the manual lookup of a path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values.
func (r *Router) Lookup(path string) (Handler, Params) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.root.getValue(path)
}
