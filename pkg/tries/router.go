package tries

import (
	"errors"
	"sync"
)

// Param consists of a key and a value.
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

// Router is a trie-based HTTP request router that efficiently handles URL path matching.
type Router struct {
	mu   sync.RWMutex
	tree node
}

// Add registers a new request handle with the given path.
// Paths must begin with '/'.
// The trailing slash is trimmed if present.
func (r *Router) Add(path string, target any) error {
	if len(path) < 2 || path[0] != '/' {
		return errors.New("path must begin with '/' in path '" + path + "'")
	}

	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tree.addRoute(path, target)
}

// Route allows the manual lookup of a path.
// This is useful to build a framework around this router.
// If the path is found, it returns the handle and path parameter values.
// If not found, it returns nil and empty Params.
func (r *Router) Route(path string) (any, Params) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tree.getValue(path)
}
