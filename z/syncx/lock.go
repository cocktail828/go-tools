package syncx

import (
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z"
)

func WithLock(locker sync.Locker, f func()) {
	locker.Lock()
	defer locker.Unlock()
	f()
}

type locker struct {
	*NamedMutex
	id any // goroutine id or name
}

func (l *locker) Lock() {
	l.holders.Store(l.id, time.Now())
	l.Locker.Lock()
}

func (l *locker) Unlock() {
	l.Locker.Unlock()
	start, ok := l.holders.LoadAndDelete(l.id)
	if ok && l.OnRelease != nil {
		l.OnRelease(l.id, time.Since(start.(time.Time)))
	}
}

type NamedMutex struct {
	sync.Locker
	OnRelease func(id any, duration time.Duration)
	holders   sync.Map // map[name]time.Time
}

// ByName returns a locker that locks by name.
func (n *NamedMutex) ByName(name string) sync.Locker {
	return &locker{
		NamedMutex: n,
		id:         name,
	}
}

// ByGoroutine returns a locker that locks by goroutine id.
func (n *NamedMutex) ByGoroutine() sync.Locker {
	return &locker{
		NamedMutex: n,
		id:         z.GoroutineID(),
	}
}
