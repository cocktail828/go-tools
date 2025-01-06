package stringx

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var (
	r     = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	chars = "abcdefghijklmnopqrstuvwxyz1234567890"
)

func RandomWidthName(width int) string {
	bytes := make([]byte, width)
	z.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}

func RandomName() string { return RandomWidthName(8) }
