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

var cfgor = &Configor{
	LoadEnv:   false,
	EnvPrefix: "",
	Unmarshal: toml.Unmarshal,
	Validator: validator.New().Struct,
}

// Load unmarshals configurations to struct from provided data.
func (c *Configor) Load(dst any, data ...[]byte) error {
	pairs := make([]pair, 0, len(data))
	for _, d := range data {
		pairs = append(pairs, pair{data: d, unmarshal: c.Unmarshal})
	}

	if err := c.internalLoad(dst, pairs...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// LoadFile unmarshals configurations to struct from provided files.
func (c *Configor) LoadFile(dst any, files ...string) error {
	pairs := make([]pair, 0, len(files))
	for _, fname := range files {
		data, err := os.ReadFile(fname)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", fname)
		}

		ext := path.Ext(fname)
		unmarshal, ok := unmarshals[ext]
		if !ok {
			return errors.Errorf("unsupported file extension: %s", ext)
		}
		pairs = append(pairs, pair{data: data, unmarshal: unmarshal})
	}

	if err := c.internalLoad(dst, pairs...); err != nil {
		return err
	}
	if c.Validator != nil {
		return c.Validator(dst)
	}
	return nil
}

// Load unmarshals configurations to struct from provided data using the default Configor.
func Load(dst any, data ...[]byte) error {
	return cfgor.Load(dst, data...)
}

// LoadFile unmarshals configurations to struct from provided files using the default Configor.
func LoadFile(dst any, files ...string) error {
	return cfgor.LoadFile(dst, files...)
}
