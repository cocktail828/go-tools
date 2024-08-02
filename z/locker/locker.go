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

func WithLockReturn[T any](locker Locker, f func() (T, error)) (T, error) {
	locker.Lock()
	defer locker.Unlock()
	return f()
}

func WithRLock(locker RWLocker, f func()) {
	locker.RLock()
	defer locker.RUnlock()
	f()
}

func WithRLockReturn[T any](locker RWLocker, f func() (T, error)) (T, error) {
	locker.RLock()
	defer locker.RUnlock()
	return f()
}
