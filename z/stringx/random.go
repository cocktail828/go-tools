package stringx

import (
	"math/rand"
	"sync"
	"time"
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
	func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	}()
	return string(bytes)
}

func RandomName() string {
	bytes := make([]byte, 8)
	func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	}()
	return string(bytes)
}
