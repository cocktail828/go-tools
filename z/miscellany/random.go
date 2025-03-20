package miscellany

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
	r        = &random{R: rand.New(rand.NewSource(time.Now().UnixNano()))}
	lowChars = "0123456789abcdefghijklmnopqrstuvwxyz"
	upChars  = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type inVariadic struct{ variadic.Assigned }
type widthKey struct{}

func WithWidth(v int) variadic.Option { return variadic.SetValue(widthKey{}, v) }
func (iv inVariadic) WithWidth() int  { return variadic.GetValue[int](iv, widthKey{}) }

type caseKey struct{}

func WithCase() variadic.Option      { return variadic.SetValue(caseKey{}, true) }
func (iv inVariadic) WithCase() bool { return variadic.GetValue[bool](iv, caseKey{}) }

// 默认长度 8, 无大写字符
func RandomName(opts ...variadic.Option) string {
	iv := inVariadic{variadic.Compose(opts...)}
	width := 8

	if w := iv.WithWidth(); w > 0 {
		width = w
	}

	chars := lowChars
	if iv.WithCase() {
		chars = upChars
	}

	bytes := make([]byte, width)
	z.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}
