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
	r       = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	upChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type option struct {
	width int
	chars string
}

type Option func(o *option)

func WithWidth(v int) Option        { return func(o *option) { o.width = v } }
func WithChars(chars string) Option { return func(o *option) { o.chars = chars } }

func RandomName(opts ...Option) string {
	o := option{
		width: 8,
		chars: upChars,
	}

	for _, opt := range opts {
		opt(&o)
	}

	bytes := make([]byte, o.width)
	z.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = o.chars[r.R.Intn(len(o.chars))]
		}
	})
	return string(bytes)
}
