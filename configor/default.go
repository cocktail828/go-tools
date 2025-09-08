package configor

import (
	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

func Load(v any, data ...[]byte) error {
	cfgor := &Configor{
		LoadEnv:      false,
		EnvPrefix:    "",
		Unmarshaller: toml.Unmarshal,
		Validator:    validator.New().Struct,
	}
	return cfgor.Load(v, data...)
}

func LoadWithUnmarshaller(v any, pairs ...Pair) error {
	cfgor := &Configor{
		LoadEnv:      false,
		EnvPrefix:    "",
		Unmarshaller: toml.Unmarshal,
		Validator:    validator.New().Struct,
	}
	return cfgor.LoadWithUnmarshaller(v, pairs...)
}
