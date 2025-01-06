package nacs

import (
	"strings"
	"sync"

	"github.com/cocktail828/go-tools/z"
)

var (
	registryMu   sync.RWMutex
	registryRepo = make(map[string]RegistryBuilder)
	configorMu   sync.RWMutex
	configorRepo = make(map[string]ConfigorBuilder)
)

type RegistryBuilder interface{ Build() Registry }
type ConfigorBuilder interface{ Build() Configor }

func Register(name string, builder any) {
	switch b := builder.(type) {
	case RegistryBuilder:
		z.WithLock(&registryMu, func() { registryRepo[strings.ToLower(name)] = b })
	case ConfigorBuilder:
		z.WithLock(&configorMu, func() { configorRepo[strings.ToLower(name)] = b })
	}
}

func LookupRegister(name string) (RegistryBuilder, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	builder, ok := registryRepo[strings.ToLower(name)]
	return builder, ok
}

func LookupConfigor(name string) (ConfigorBuilder, bool) {
	configorMu.RLock()
	defer configorMu.RUnlock()
	builder, ok := configorRepo[strings.ToLower(name)]
	return builder, ok
}
