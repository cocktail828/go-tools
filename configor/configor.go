package configor

import (
	"os"
	"reflect"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type Configor struct {
	LoadEnv      bool                    // 是否读取环境变量
	EnvPrefix    string                  // 环境变量前缀
	Unmarshaller func([]byte, any) error // 解析器
	Validator    func(any) error         // 校验器
}

type Pair struct {
	data         []byte
	unmarshaller func([]byte, any) error
}

func (c *Configor) bindEnv(in any) error {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.Errorf("input must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if field.Elem().Kind() == reflect.Ptr {
				return errors.Errorf("unsupported nested pointer type: %s", field.Type())
			}
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := c.bindEnv(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		envName := structField.Tag.Get("env")
		if envName != "" && envName != "-" && c.EnvPrefix != "" {
			envName = c.EnvPrefix + "_" + envName
		}

		envVal := ""
		if c.LoadEnv && envName != "" && envName != "-" {
			envVal = os.Getenv(envName)
		}

		if envVal == "" {
			envVal = structField.Tag.Get("default")
		}

		if envVal == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(envVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(envVal, 10, 64)
			if err != nil {
				return errors.Errorf("error parsing %s as int64: %v", envName, err)
			}
			field.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			intValue, err := strconv.ParseUint(envVal, 10, 64)
			if err != nil {
				return errors.Errorf("error parsing %s as uint64: %v", envName, err)
			}
			field.SetUint(intValue)
		case reflect.Float32:
			floatValue, err := strconv.ParseFloat(envVal, 32)
			if err != nil {
				return errors.Errorf("error parsing %s as float32: %v", envName, err)
			}
			field.SetFloat(floatValue)
		case reflect.Float64:
			floatValue, err := strconv.ParseFloat(envVal, 64)
			if err != nil {
				return errors.Errorf("error parsing %s as float64: %v", envName, err)
			}
			field.SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(envVal)
			if err != nil {
				return errors.Errorf("error parsing %s as bool: %v", envName, err)
			}
			field.SetBool(boolValue)
		default:
			// return errors.Errorf("unsupported field type: %s", field.Kind())
		}
	}
	return nil
}

func (c *Configor) Load(v any, data ...[]byte) error {
	pairs := make([]Pair, 0, len(data))
	for _, d := range data {
		pairs = append(pairs, Pair{d, c.Unmarshaller})
	}
	return c.LoadWithUnmarshaller(v, pairs...)
}

func (c *Configor) LoadWithUnmarshaller(v any, pairs ...Pair) error {
	for i, d := range pairs {
		if d.unmarshaller == nil {
			return errors.Errorf("unmarshaller is nil at index %d", i)
		}
	}

	if riv := reflect.Indirect(reflect.ValueOf(v)); !riv.CanAddr() {
		return errors.Errorf("target %v must be addressable", v)
	}

	if err := c.bindEnv(v); err != nil {
		return err
	}

	for _, p := range pairs {
		if err := p.unmarshaller(p.data, v); err != nil {
			return err
		}
	}

	if c.Validator != nil {
		return c.Validator(v)
	}
	return nil
}

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
