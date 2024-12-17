package locker

import (
	"sync"
)

type Locker sync.Locker
type RWLocker interface {
	RLock()
	RUnlock()
}

func WithLock(locker Locker, f func()) {
	locker.Lock()
	defer locker.Unlock()
	f()
}

func WithRLock(locker RWLocker, f func()) {
	locker.RLock()
	defer locker.RUnlock()
	f()
}
