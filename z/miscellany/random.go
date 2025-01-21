package miscellany

import (
	"math/rand"
	"sync"
	"time"
	"unicode"

	"github.com/cocktail828/go-tools/z"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var (
	r     = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	chars = "0123456789abcdefghijklmnopqrstuvwxyz"
)

type inOption struct {
	width    int
	withCase bool
}

type option func(*inOption)

func WithWidth(n int) option {
	return func(io *inOption) {
		io.width = n
	}
}

func WithCase() option {
	return func(io *inOption) {
		io.withCase = true
	}
}

func RandomName(opts ...option) string {
	o := inOption{width: 8}
	for _, f := range opts {
		f(&o)
	}

	bytes := make([]byte, o.width)
	z.WithLock(r, func() {
		for i := range bytes {
			char := chars[r.R.Intn(len(chars))]
			if o.withCase {
				if rn := r.R.Intn(100); rn > 50 {
					char = byte(unicode.ToUpper(rune(char)))
				}
			}
			bytes[i] = char
		}
	})
	return string(bytes)
}
