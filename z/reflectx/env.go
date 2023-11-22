package reflectx

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func reflectSetVal(rv reflect.Value, val string) error {
	if ok := rv.CanSet(); !ok {
		return errors.New("target value is not settable")
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(val)
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		rv.SetBool(v)
	case reflect.Float32:
		v, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		rv.SetFloat(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(v)
	case reflect.Int8:
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Int16:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Int32:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Int:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		rv.SetInt(v)
	case reflect.Uint8:
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		rv.SetUint(v)
	case reflect.Uint16:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		rv.SetUint(v)
	case reflect.Uint32:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		rv.SetUint(v)
	case reflect.Uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(v)
	case reflect.Uint:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		rv.SetUint(v)
	default:
		return errors.Errorf("unsupport type: %v", rv.Kind())
	}
	return nil
}

func BindEnvVal(target interface{}, name string) error {
	val := os.Getenv(name)
	if val == "" {
		return errors.Errorf("env '%v' is not found or unset", name)
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("target should be a non-nil pointer")
	}
	return reflectSetVal(rv.Elem(), val)
}

func BindEnv(target interface{}) error {
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
		arr := []string{}
		if tag := rtype.Field(pos).Tag.Get("env"); tag != "" {
			arr = strings.Split(tag, ",")
		}
		key := rtype.Field(pos).Name
		opt := "optional"
		switch {
		case len(arr) == 1:
			key = arr[0]
		case len(arr) >= 2:
			key, opt = arr[0], arr[1]
		}

		eval := os.Getenv(key)
		if opt == "required" && eval == "" {
			return errors.Errorf("env '%v' is required but not found", key)
		}

		if eval == "" {
			eval = rtype.Field(pos).Tag.Get("default")
		}
		if eval != "" {
			if err := reflectSetVal(rvalue.Field(pos), eval); err != nil {
				return err
			}
		}
	}
	return nil
}
