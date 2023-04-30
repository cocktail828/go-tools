package locker

import (
	"sync"
)

type Locker sync.Locker
type RWLocker interface {
	Locker
	RLock()
	RUnlock()
}

func WithLock[T interface{}](locker Locker, f func() T) T {
	locker.Lock()
	defer locker.Unlock()
	return f()
}

func WithRLock[T interface{}](locker RWLocker, f func() T) T {
	locker.RLock()
	defer locker.RUnlock()
	return f()
}
