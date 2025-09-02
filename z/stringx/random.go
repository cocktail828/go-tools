package stringx

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/cocktail828/go-tools/z/variadic"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var (
	r       = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	upChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type widthKey struct{}

func WithWidth(v int) variadic.Option           { return variadic.Set(widthKey{}, v) }
func getWidth(c variadic.Container) (int, bool) { return variadic.Get[int](c, widthKey{}) }

type charsKey struct{}

func WithChars(chars string) variadic.Option       { return variadic.Set(charsKey{}, chars) }
func getChars(c variadic.Container) (string, bool) { return variadic.Get[string](c, charsKey{}) }

// 默认长度 8
func RandomName(opts ...variadic.Option) string {
	iv := variadic.Compose(opts...)

	width := 8
	if w, ok := getWidth(iv); ok && w > 0 {
		width = w
	}

	chars := upChars
	if s, ok := getChars(iv); ok && s != "" {
		chars = s
	}

	bytes := make([]byte, width)
	z.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}
