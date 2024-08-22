package stringx

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z/locker"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var (
	r     = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	chars = "abcdefghijklmnopqrstuvwxyz"
)

func RandomWidthName(width int) string {
	bytes := make([]byte, width)
	locker.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}

func RandomName() string { return RandomWidthName(8) }
