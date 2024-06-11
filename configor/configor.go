package configor

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

type Configor struct {
	EnvPrefix   string
	Unmarshaler func([]byte, any) error
	Validator   func(any) error
}

// New initialize a Configor
func newConfigor() *Configor {
	return &Configor{
		EnvPrefix:   strings.ToUpper(os.Getenv("CONFIGOR_ENV_PREFIX")),
		Unmarshaler: toml.Unmarshal,
		Validator:   validator.New().Struct,
	}
}

func (c *Configor) Load(dst any, payload ...[]byte) (err error) {
	if err := c.load(dst, payload...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// Load will unmarshal configurations to struct from files that you provide
func (c *Configor) LoadFile(dst any, files ...string) error {
	if err := c.loadFile(dst, files...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// Load will unmarshal configurations to struct from files that you provide
func Load(dst any, payload ...[]byte) error {
	return newConfigor().Load(dst, payload...)
}

// Load will unmarshal configurations to struct from files that you provide
func LoadFile(dst any, files ...string) error {
	return newConfigor().LoadFile(dst, files...)
}
