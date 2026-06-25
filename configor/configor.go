package configor

import (
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Configor struct {
	LoadEnv      bool                    // 是否读取环境变量
	EnvPrefix    string                  // 环境变量前缀
	Unmarshaller func([]byte, any) error // 解析器
	Validator    func(any) error         // 校验器
}

type Pair struct {
	Data         []byte
	Unmarshaller func([]byte, any) error
}

func (c *Configor) setFieldValue(field reflect.Value, name string, val string) error {
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		durationValue, err := time.ParseDuration(val)
		if err != nil {
			return errors.Errorf("error parsing %s as time.Duration: %v", name, err)
		}
		field.Set(reflect.ValueOf(durationValue))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return errors.Errorf("error parsing %s as int64: %v", name, err)
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intValue, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return errors.Errorf("error parsing %s as uint64: %v", name, err)
		}
		field.SetUint(intValue)
	case reflect.Float32:
		floatValue, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return errors.Errorf("error parsing %s as float32: %v", name, err)
		}
		field.SetFloat(floatValue)
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return errors.Errorf("error parsing %s as float64: %v", name, err)
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(val)
		if err != nil {
			return errors.Errorf("error parsing %s as bool: %v", name, err)
		}
		field.SetBool(boolValue)
	}
	return nil
}

func (c *Configor) walkFields(in any, fn func(field reflect.Value, structField reflect.StructField) error) error {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.Errorf("input must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		if !field.CanSet() {
			continue
		}

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
			if err := c.walkFields(field.Addr().Interface(), fn); err != nil {
				return err
			}
			continue
		}

		if err := fn(field, structField); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configor) bindDefault(in any) error {
	return c.walkFields(in, func(field reflect.Value, structField reflect.StructField) error {
		defVal := structField.Tag.Get("default")
		if defVal == "" {
			return nil
		}
		return c.setFieldValue(field, structField.Name, defVal)
	})
}

func (c *Configor) bindEnv(in any) error {
	if !c.LoadEnv {
		return nil
	}
	return c.walkFields(in, func(field reflect.Value, structField reflect.StructField) error {
		envName := structField.Tag.Get("env")
		if envName == "" || envName == "-" {
			return nil
		}
		if c.EnvPrefix != "" {
			envName = c.EnvPrefix + "_" + envName
		}
		envVal := os.Getenv(envName)
		if envVal == "" {
			return nil
		}
		return c.setFieldValue(field, envName, envVal)
	})
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
		if d.Unmarshaller == nil {
			return errors.Errorf("unmarshaller is nil at index %d", i)
		}
	}

	if riv := reflect.Indirect(reflect.ValueOf(v)); !riv.CanAddr() {
		return errors.Errorf("target %v must be addressable", v)
	}

	// Priority: env > file > default
	if err := c.bindDefault(v); err != nil {
		return err
	}

	for _, p := range pairs {
		if err := p.Unmarshaller(p.Data, v); err != nil {
			return err
		}
	}

	if err := c.bindEnv(v); err != nil {
		return err
	}

	if c.Validator != nil {
		return c.Validator(v)
	}
	return nil
}
