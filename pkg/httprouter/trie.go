package httprouter

import (
	"net/http"
	"sync"
)

type Task func(p Params)

type Trie struct {
	mu             sync.RWMutex
	root           node
	NotFoundHandle Task
}

func (t *Trie) AddRoute(path string, task Task) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root.addRoute(path, func(w http.ResponseWriter, r *http.Request, p Params) {
		task(p)
	})
}

func (t *Trie) AddRoutes(routes map[string]Task) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for path, task := range routes {
		t.root.addRoute(path, func(w http.ResponseWriter, r *http.Request, p Params) {
			task(p)
		})
	}
}

// Handle 处理请求路径，调用注册的处理函数
func (t *Trie) Handle(path string) {
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	handle, p, _ := t.root.getValue(path)
	if handle == nil {
		if t.NotFoundHandle != nil {
			t.NotFoundHandle(p)
		}
		return
	}

	handle(nil, nil, p)
}
