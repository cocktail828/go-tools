package configor

import (
	"encoding/json"
	stderr "errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func LoadFile(dst any, files ...string) error {
	return newConfigor().LoadFile(dst, files...)
}

func Load(dst any, data ...[]byte) error {
	return newConfigor().Load(dst, data...)
}

type Configor struct {
	// default unmarshaler for raw data
	Unmarshaler func([]byte, any) error
}

func newConfigor() *Configor {
	return &Configor{
		Unmarshaler: defaultUnmarshaler,
	}
}

// Load will unmarshal configurations to struct from files that you provide
func (c *Configor) LoadFile(dst any, files ...string) error {
	for _, fname := range files {
		if fileInfo, err := os.Stat(fname); err != nil || fileInfo.Mode().IsRegular() {
			return errors.Errorf("no such fname: %v", fname)
		}

		data, err := os.ReadFile(fname)
		if err != nil {
			return err
		}

		f, ok := unmarshalers[path.Ext(fname)]
		if !ok {
			return errors.Errorf("no such unmarshaler for:%v", path.Ext(fname))
		}
		if err := f(data, dst); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}

func (c *Configor) Load(dst any, data ...[]byte) error {
	for _, d := range data {
		if err := c.Unmarshaler(d, dst); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}

type handler = func(fieldStruct reflect.StructField, field reflect.Value) error

func handleEnv(fieldStruct reflect.StructField, field reflect.Value) error {
	envNames := []string{fieldStruct.Name, strings.ToUpper(fieldStruct.Name)}
	if envName := fieldStruct.Tag.Get("env"); envName != "" {
		envNames = []string{envName}
	}

	errs := []error{}
	for _, env := range envNames {
		if eval := os.Getenv(env); eval != "" {
			if err := json.Unmarshal([]byte(eval), field.Addr().Interface()); err == nil {
				return nil
			} else {
				fmt.Println(err)
				errs = append(errs, err)
			}
		}
	}
	return stderr.Join(errs...)
}

func handleDefault(fieldStruct reflect.StructField, field reflect.Value) error {
	if dflt := fieldStruct.Tag.Get("default"); dflt != "" {
		if err := json.Unmarshal([]byte(dflt), field.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configor) processTags(dst any) error {
	rval := reflect.Indirect(reflect.ValueOf(dst))
	if rval.Kind() != reflect.Struct {
		return errors.New("invalid arg, should be struct")
	}

	funcs := []handler{handleEnv, handleDefault}
	types := rval.Type()
	for i := 0; i < types.NumField(); i++ {
		fieldStruct := types.Field(i)
		field := rval.Field(i)
		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		for _, f := range funcs {
			if !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
				break
			}
			if err := f(fieldStruct, field); err != nil {
				return err
			}
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			if strings.ToLower(fieldStruct.Tag.Get("required")) == "true" {
				return errors.Errorf("Field %v is required, but blank", fieldStruct.Name)
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := c.processTags(field.Addr().Interface()); err != nil {
				return err
			}

		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := c.processTags(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
