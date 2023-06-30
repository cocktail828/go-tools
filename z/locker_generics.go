//go:build go1.18
// +build go1.18

package z

func WithGenericsLock[T interface{}](locker Locker, f func() T) T {
	locker.Lock()
	defer locker.Unlock()
	return f()
}

func WithGenericsRLock[T interface{}](locker RWLocker, f func() T) T {
	locker.RLock()
	defer locker.RUnlock()
	return f()
}
