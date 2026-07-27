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

type params struct {
	rnd           *rand.Rand
	width         int    // default 8
	chars         string // default AlphabetDigets
	isCapitalized bool
}

type Option func(o *params)

func WithWidth(v int) Option             { return func(o *params) { o.width = v } }
func WithChars(chars string) Option      { return func(o *params) { o.chars = chars } }
func WithRandomizer(r *rand.Rand) Option { return func(o *params) { o.rnd = r } }
func WithCapitalized() Option            { return func(o *params) { o.isCapitalized = true } }

// RandomName returns a random name with the given options.
// If width is not specified, it will default to 8.
// If chars is not specified, it will default to AlphabetDigets.
func RandomName(opts ...Option) string {
	o := params{
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

	s := sb.String()
	if len(s) > 1 && o.isCapitalized {
		return strings.ToUpper(s[:1]) + s[1:]
	}

	return s
}
