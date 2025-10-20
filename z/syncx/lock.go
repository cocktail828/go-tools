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
	id     any       // goroutine id or name
	holdAt time.Time // the beging time of get lock or blocked by a lock
}

func (l *locker) Lock() {
	if l.OnRelease != nil {
		l.holdAt = time.Now()
	}
	l.Locker.Lock()
}

func (l *locker) Unlock() {
	l.Locker.Unlock()
	if l.OnRelease != nil {
		if l.OnRelease != nil && !l.holdAt.IsZero() {
			l.OnRelease(l.id, time.Since(l.holdAt))
		}
	}
}

type NamedMutex struct {
	sync.Locker
	OnRelease func(id any, duration time.Duration)
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
