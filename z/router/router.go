package router

import (
	"sync"

	"github.com/cocktail828/go-tools/httprouter/trie"
)

type Handler func(path string, ps trie.Params)

// Router is a Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	mu   sync.RWMutex
	root *trie.Node
}

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{root: new(trie.Node)}
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
	r.root.AddRoute(path, handle)
}

// Lookup allows the manual lookup of a path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values.
func (r *Router) Lookup(path string) (Handler, trie.Params) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, p, _ := r.root.GetValue(path)
	if h == nil {
		return nil, p
	}
	return h.(Handler), p
}
