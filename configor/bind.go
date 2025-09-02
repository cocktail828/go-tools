package configor

import (
	"os"
	"reflect"
	"strconv"

	"github.com/cocktail828/go-tools/z/variadic"
	"github.com/pkg/errors"
)

type inVariadic struct{ variadic.Assigned }
type prefixKey struct{}

func WithPrefix(v string) variadic.Option { return variadic.SetValue(prefixKey{}, v) }
func (iv inVariadic) WithPrefix() string  { return variadic.GetValue[string](iv, prefixKey{}) }

type skipenvKey struct{}

func WithSkipEnv(v bool) variadic.Option { return variadic.SetValue(skipenvKey{}, v) }
func (iv inVariadic) WithSkipEnv() bool  { return variadic.GetValue[bool](iv, skipenvKey{}) }

func BindEnv(in any, opts ...variadic.Option) error {
	biv := inVariadic{variadic.Compose(opts...)}
	prefix := biv.WithPrefix()
	skipEnv := biv.WithSkipEnv()

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
			if err := BindEnv(field.Addr().Interface(), opts...); err != nil {
				return err
			}
			continue
		}

		envName := structField.Tag.Get("env")
		if envName != "" && envName != "-" && prefix != "" {
			envName = prefix + "_" + envName
		}

		envVal := ""
		if !skipEnv {
			envVal = os.Getenv(envName)
		}

		if skipEnv || envVal == "" {
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
