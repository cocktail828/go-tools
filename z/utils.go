package z

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

func TryPut[T chan E, E any](c T, e E) bool {
	select {
	case c <- e:
		return true
	default:
		return false
	}
}

func TryGet[T chan E, E any](c T) (E, bool) {
	select {
	case e, ok := <-c:
		if ok {
			return e, true
		}
	default:
	}

	var e E
	return e, false
}
