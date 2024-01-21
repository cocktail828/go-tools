package configor

import (
	"encoding/json"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

var (
	defaultUnmarshaler = toml.Unmarshal
	unmarshalers       = map[string]func([]byte, any) error{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".toml": toml.Unmarshal,
		".json": json.Unmarshal,
	}
)

func Register(suffix string, f func([]byte, any) error) {
	unmarshalers[suffix] = f
}

func Unmarshalers() []string {
	r := []string{}
	for k := range unmarshalers {
		r = append(r, k)
	}
	return r
}
