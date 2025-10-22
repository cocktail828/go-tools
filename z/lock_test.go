package z

import (
	"sync"
	"testing"
	"time"
)

func TestNamedMutex(t *testing.T) {
	nmu := &NamedMutex{
		Locker: &sync.Mutex{},
		OnRelease: func(id any, duration time.Duration) {
			t.Logf("unlock %v, cost %v", id, duration)
		},
	}

	mu1 := nmu.ByName("xxx")
	mu2 := nmu.ByName("xxx")
	mu1.Lock()
	time.Sleep(time.Millisecond * 100)
	defer mu2.Unlock()
}
