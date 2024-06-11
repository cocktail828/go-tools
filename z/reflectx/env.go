package reflectx

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// BindEnv binds environment variables to struct fields based on 'env' tags
func BindEnv(in interface{}) error {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("input must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if envName := structField.Tag.Get("env"); envName != "" {
			if envValue := os.Getenv(envName); envValue != "" {
				switch field.Kind() {
				case reflect.String:
					field.SetString(envValue)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					intValue, err := strconv.ParseInt(envValue, 10, 64)
					if err != nil {
						return fmt.Errorf("error parsing %s as int64: %v", envName, err)
					}
					field.SetInt(intValue)
				case reflect.Float32, reflect.Float64:
					floatValue, err := strconv.ParseFloat(envValue, 32)
					if err != nil {
						return fmt.Errorf("error parsing %s as float32: %v", envName, err)
					}
					field.SetFloat(floatValue)
				default:
					return fmt.Errorf("unsupported field type: %s", field.Kind())
				}
			}
		}

		if field.Kind() == reflect.Struct {
			err := BindEnv(field.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
