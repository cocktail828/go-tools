package chain

import (
	"fmt"
	"strings"
	"sync"
)

type Handler interface {
	Name() string
}

type ResponsibeChain struct {
	mu         *sync.RWMutex
	handlerMap map[string]struct{}
	handlers   []Handler
}

func New() *ResponsibeChain {
	return &ResponsibeChain{
		mu:         &sync.RWMutex{},
		handlerMap: make(map[string]struct{}),
	}
}

func (rc ResponsibeChain) String() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return fmt.Sprintf("handlers: [%v]", func() string {
		strs := []string{}
		for _, handler := range rc.handlers {
			strs = append(strs, handler.Name())
		}
		return strings.Join(strs, ",")
	}())
}

func (rc *ResponsibeChain) Reset() *ResponsibeChain {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.handlerMap = make(map[string]struct{})
	rc.handlers = rc.handlers[:0]
	return rc
}

func (rc *ResponsibeChain) Register(h Handler) *ResponsibeChain {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if _, ok := rc.handlerMap[h.Name()]; ok {
		rc.handlerMap[h.Name()] = struct{}{}
		rc.handlers = append(rc.handlers, h)
	}
	return rc
}

func (rc *ResponsibeChain) Length() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.handlers)
}

func (rc *ResponsibeChain) Exists(name string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	_, ok := rc.handlerMap[name]
	return ok
}

func (rc *ResponsibeChain) Traverse(f func(h Handler) bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	for _, h := range rc.handlers {
		if !f(h) {
			break
		}
	}
}
