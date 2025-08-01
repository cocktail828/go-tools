package configuration

import (
	"strings"
	"sync"

	"github.com/cocktail828/go-tools/z"
)

var (
	configorMu   sync.RWMutex
	configorRepo = make(map[string]Builder)
)

type Builder interface{ Build() Configor }

func Register(name string, b Builder) {
	z.WithLock(&configorMu, func() {
		configorRepo[strings.ToLower(name)] = b
	})
}

func Lookup(name string) (Builder, bool) {
	configorMu.RLock()
	defer configorMu.RUnlock()
	builder, ok := configorRepo[strings.ToLower(name)]
	return builder, ok
}
