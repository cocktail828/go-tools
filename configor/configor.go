package configor

import (
	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

type Configor struct {
	LoadEnv   bool            // 是否读取环境变量
	EnvPrefix string          // 环境变量前缀
	Unmarshal Unmarshal       // 解析器
	Validator func(any) error // 校验器
}

var Default = &Configor{
	LoadEnv:   false,
	EnvPrefix: "",
	Unmarshal: toml.Unmarshal,
	Validator: validator.New().Struct,
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
	return Default.Load(dst, payload...)
}

// Load will unmarshal configurations to struct from files that you provide
func LoadFile(dst any, files ...string) error {
	return Default.LoadFile(dst, files...)
}
