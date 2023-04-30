package z

import (
	"math/rand"
	"sync"
	"time"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var r = &random{
	R: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func GenerateRandomName() string {
	r.Lock()
	defer r.Unlock()
	chars := "abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 8)
	for i := range bytes {
		bytes[i] = chars[r.R.Intn(len(chars))]
	}
	return string(bytes)
}
