package environ

import (
	"os"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

func BindEnv(in interface{}) error {
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
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := BindEnv(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		envName := structField.Tag.Get("env")
		if envName == "" || envName == "-" {
			continue
		}

		envValue := os.Getenv(envName)
		if envValue == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return errors.Errorf("error parsing %s as int64: %v", envName, err)
			}
			field.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			intValue, err := strconv.ParseUint(envValue, 10, 64)
			if err != nil {
				return errors.Errorf("error parsing %s as uint64: %v", envName, err)
			}
			field.SetUint(intValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(envValue, 32)
			if err != nil {
				return errors.Errorf("error parsing %s as float32: %v", envName, err)
			}
			field.SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(envValue)
			if err != nil {
				return errors.Errorf("error parsing %s as bool: %v", envName, err)
			}
			field.SetBool(boolValue)
		default:
			return errors.Errorf("unsupported field type: %s", field.Kind())
		}
	}
	return nil
}
