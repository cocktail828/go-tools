package configor

import (
	"reflect"

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

func Load(dst any, data ...[]byte) error {
	return cfgor.Load(dst, data...)
}

type Unmarshal func([]byte, any) error

type pair struct {
	data      []byte
	unmarshal Unmarshal
}

func (c *Configor) internalLoad(dst any, pairs ...pair) error {
	defaultValue := reflect.Indirect(reflect.ValueOf(dst))
	if !defaultValue.CanAddr() {
		return errors.Errorf("config %v must be addressable", dst)
	}

	if err := BindEnv(dst, WithPrefix(c.EnvPrefix), WithSkipEnv(!c.LoadEnv)); err != nil {
		return errors.Wrap(err, "fail to process env or defaults")
	}

	for _, val := range pairs {
		if err := val.unmarshal(val.data, dst); err != nil {
			return errors.Wrap(err, "fail to unmarshal data")
		}
	}
	return nil
}
