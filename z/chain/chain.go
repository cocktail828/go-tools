package chain

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type Handler interface {
	Name() string
}

type Chain struct {
	mu         *sync.RWMutex
	handlerMap map[string]*list.Element
	handlers   *list.List
}

func New() *Chain {
	return &Chain{
		mu:         &sync.RWMutex{},
		handlerMap: make(map[string]*list.Element),
		handlers:   list.New(),
	}
}

func (rc Chain) String() string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return fmt.Sprintf("handlers: [%v]", func() string {
		strs := []string{}
		for h := rc.handlers.Front(); h != nil; h = h.Next() {
			strs = append(strs, h.Next().Value.(Handler).Name())
		}
		return strings.Join(strs, ",")
	}())
}

func (rc *Chain) Reset() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for _, h := range rc.handlerMap {
		rc.handlers.Remove(h)
	}
	rc.handlerMap = make(map[string]*list.Element)
}

func (rc *Chain) Add(hs ...Handler) *Chain {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for _, h := range hs {
		if _, ok := rc.handlerMap[h.Name()]; !ok {
			rc.handlerMap[h.Name()] = rc.handlers.PushBack(h)
		}
	}
	return rc
}

func (rc *Chain) Remove(names ...string) *Chain {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for _, name := range names {
		if h, ok := rc.handlerMap[name]; ok {
			delete(rc.handlerMap, name)
			rc.handlers.Remove(h)
		}
	}
	return rc
}

func (rc *Chain) Get(name string) Handler {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if h, ok := rc.handlerMap[name]; ok {
		return h.Value.(Handler)
	}
	return nil
}

func (rc *Chain) Len() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.handlers.Len()
}

func (rc *Chain) Traverse(e *list.Element, f func(h Handler) bool) *list.Element {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if e == nil {
		e = rc.handlers.Front()
	}
	for h := e; h != nil; h = h.Next() {
		if !f(h.Value.(Handler)) {
			return h
		}
	}
	return nil
}

func (rc *Chain) Reverse(e *list.Element, f func(h Handler) bool) *list.Element {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if e == nil {
		e = rc.handlers.Back()
	}
	for h := e; h != nil; h = h.Prev() {
		if !f(h.Value.(Handler)) {
			return h
		}
	}
	return nil
}

func (rc *Chain) Clone() *Chain {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	c := New()
	for h := rc.handlers.Front(); h != nil; h = h.Next() {
		v := h.Value.(Handler)
		c.handlerMap[v.Name()] = c.handlers.PushBack(v)
	}
	return c
}
