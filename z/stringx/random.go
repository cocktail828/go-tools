package stringx

import (
	"math/rand"
	"strings"
	"time"
)

var (
	Digets          = "0123456789"
	AlphabetLowCase = "abcdefghijklmnopqrstuvwxyz"
	AlphabetUpCase  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Alphabet        = AlphabetLowCase + AlphabetUpCase
	AlphabetDigets  = Digets + Alphabet
)

type option struct {
	*rand.Rand
	width int    // default 8
	chars string // default AlphabetDigets
}

type Option func(o *option)

func WithWidth(v int) Option             { return func(o *option) { o.width = v } }
func WithChars(chars string) Option      { return func(o *option) { o.chars = chars } }
func WithRandomizer(r *rand.Rand) Option { return func(o *option) { o.Rand = r } }

// RandomName returns a random name with the given options.
// If width is not specified, it will default to 8.
// If chars is not specified, it will default to AlphabetDigets.
func RandomName(opts ...Option) string {
	o := option{
		width: 8,
		chars: AlphabetDigets,
		Rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, opt := range opts {
		opt(&o)
	}

	sb := strings.Builder{}
	for range o.width {
		sb.WriteByte(o.chars[o.Intn(len(o.chars))])
	}
	return sb.String()
}
