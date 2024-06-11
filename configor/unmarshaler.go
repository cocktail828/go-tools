package configor

import (
	"encoding/json"
	"sync"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

var (
	unmarshalerMu sync.RWMutex
	unmarshalers  = map[string]func([]byte, any) error{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".toml": toml.Unmarshal,
		".json": json.Unmarshal,
	}
)

func Register(suffix string, f func([]byte, any) error) {
	unmarshalerMu.Lock()
	defer unmarshalerMu.Unlock()
	unmarshalers[suffix] = f
}

func Unmarshalers() []string {
	unmarshalerMu.RLock()
	defer unmarshalerMu.RUnlock()
	r := []string{}
	for k := range unmarshalers {
		r = append(r, k)
	}
	return r
}
