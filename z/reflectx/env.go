package reflectx

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

func BindEnv(target interface{}) error {
	v := viper.New()
	v.AutomaticEnv()

	rtype := reflect.TypeOf(target)
	rvalue := reflect.ValueOf(target)
	if rvalue.Kind() != reflect.Pointer || rvalue.IsNil() {
		return errors.New("target should be a non-nil pointer")
	}
	rtype = rtype.Elem()
	rvalue = rvalue.Elem()
	if rvalue.Kind() != reflect.Struct {
		return errors.New("target should be a non-nil struct pointer")
	}

	for pos := 0; pos < rtype.NumField(); pos++ {
		envname := rtype.Field(pos).Tag.Get("env")
		envval := rtype.Field(pos).Tag.Get("default")
		if envval == "" {
			continue
		}
		if envname == "" {
			envname = rtype.Field(pos).Name
		}
		os.Setenv(strings.ToUpper(envname), envval)
	}

	if err := v.Unmarshal(target); err != nil {
		return err
	}

	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		if val := field.Tag.Get("env"); val != "" {
			return val
		}
		return field.Name
	})
	return validate.Struct(target)
}
