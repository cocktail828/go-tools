package z

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z/locker"
	"github.com/cocktail828/go-tools/z/reflectx"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var r = &random{
	R: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func GenerateRandomName() string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 8)

	locker.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}

func Must(err error) {
	if !reflectx.IsNil(err) {
		panic(err)
	}
}
