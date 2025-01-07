package configor

import (
	"encoding/json"
	"sync"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Unmarshal func([]byte, any) error

var (
	unmarshalMu sync.RWMutex
	unmarshals  = map[string]Unmarshal{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".toml": toml.Unmarshal,
		".json": json.Unmarshal,
	}
)

// Register registers a new unmarshal function for a file extension.
func Register(suffix string, f Unmarshal) {
	unmarshalMu.Lock()
	defer unmarshalMu.Unlock()
	unmarshals[suffix] = f
}
