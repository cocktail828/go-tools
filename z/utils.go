package z

import "sync"

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

func Contains[E comparable](s []E, e E) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
