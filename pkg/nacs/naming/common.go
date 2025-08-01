package naming

import (
	"strings"
	"sync"

	"github.com/cocktail828/go-tools/z"
)

var (
	registryMu   sync.RWMutex
	registryRepo = make(map[string]Builder)
)

type Builder interface{ Build() Registry }

func Register(name string, b Builder) {
	z.WithLock(&registryMu, func() {
		registryRepo[strings.ToLower(name)] = b
	})
}

func Lookup(name string) (Builder, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	builder, ok := registryRepo[strings.ToLower(name)]
	return builder, ok
}

type Endpoint struct {
	IP   string
	Port int
}
