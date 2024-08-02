package configor

import (
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
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

func (c *Configor) Load(dst any, data ...[]byte) (err error) {
	pairs := make([]pair, 0, len(data))
	for idx := 0; idx < len(data); idx++ {
		pairs = append(pairs, pair{data: data[idx], unmarshal: c.Unmarshal})
	}

	if err := c.internalLoad(dst, pairs...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// Load will unmarshal configurations to struct from files that you provide
func (c *Configor) LoadFile(dst any, files ...string) error {
	pairs := make([]pair, 0, len(files))
	for _, fname := range files {
		data, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		if f, ok := unmarshals[path.Ext(fname)]; ok {
			pairs = append(pairs, pair{data, f})
		} else {
			return errors.Errorf("missing unmarshal for %q", path.Ext(fname))
		}
	}

	if err := c.internalLoad(dst, pairs...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// Load will unmarshal configurations to struct from files that you provide
func Load(dst any, data ...[]byte) error {
	return Default.Load(dst, data...)
}

// Load will unmarshal configurations to struct from files that you provide
func LoadFile(dst any, files ...string) error {
	return Default.LoadFile(dst, files...)
}
