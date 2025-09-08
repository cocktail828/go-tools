package syncx

import (
	"sync"
	"time"
)

func WithLock(locker sync.Locker, f func()) {
	locker.Lock()
	defer locker.Unlock()
	f()
}

type NamedMutex struct {
	sync.Locker
	OnRelease func(name string, duration time.Duration)
	holders   sync.Map // map[name]time.Time
}

func (n *NamedMutex) Lock(name string) {
	n.holders.Store(name, time.Now())
	n.Locker.Lock()
}

func (n *NamedMutex) Unlock(name string) {
	n.Locker.Unlock()
	start, ok := n.holders.LoadAndDelete(name)
	if ok && n.OnRelease != nil {
		n.OnRelease(name, time.Since(start.(time.Time)))
	}
}
