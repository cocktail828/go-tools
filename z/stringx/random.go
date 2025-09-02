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

type inVariadic struct{ variadic.Assigned }
type widthKey struct{}

func WithWidth(v int) variadic.Option { return variadic.SetValue(widthKey{}, v) }
func (iv inVariadic) WithWidth() int  { return variadic.GetValue[int](iv, widthKey{}) }

type charsKey struct{}

func WithChars(chars string) variadic.Option { return variadic.SetValue(charsKey{}, chars) }
func (iv inVariadic) WithChars() string      { return variadic.GetValue[string](iv, charsKey{}) }

// 默认长度 8, 无大写字符
func RandomName(opts ...variadic.Option) string {
	iv := inVariadic{variadic.Compose(opts...)}
	width := 8

	if w := iv.WithWidth(); w > 0 {
		width = w
	}

	chars := upChars
	if s := iv.WithChars(); s != "" {
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
