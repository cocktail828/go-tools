package stringx

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	Digets         = "0123456789"
	AlphabetLower  = "abcdefghijklmnopqrstuvwxyz"
	AlphabetUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Alphabet       = AlphabetLower + AlphabetUpper
	AlphabetDigets = Digets + Alphabet
)

var globalRand = struct {
	sync.Mutex
	*rand.Rand
}{Rand: rand.New(rand.NewSource(time.Now().UnixNano()))}

type option struct {
	rnd   *rand.Rand
	width int    // default 8
	chars string // default AlphabetDigets
}

type Option func(o *option)

func WithWidth(v int) Option             { return func(o *option) { o.width = v } }
func WithChars(chars string) Option      { return func(o *option) { o.chars = chars } }
func WithRandomizer(r *rand.Rand) Option { return func(o *option) { o.rnd = r } }

// RandomName returns a random name with the given options.
// If width is not specified, it will default to 8.
// If chars is not specified, it will default to AlphabetDigets.
func RandomName(opts ...Option) string {
	o := option{
		width: 8,
		chars: AlphabetDigets,
	}

	for _, opt := range opts {
		opt(&o)
	}

	sb := strings.Builder{}
	sb.Grow(o.width)

	if o.rnd != nil {
		for range o.width {
			sb.WriteByte(o.chars[o.rnd.Intn(len(o.chars))])
		}
	} else {
		globalRand.Lock()
		for range o.width {
			sb.WriteByte(o.chars[globalRand.Intn(len(o.chars))])
		}
		globalRand.Unlock()
	}
	return sb.String()
}
