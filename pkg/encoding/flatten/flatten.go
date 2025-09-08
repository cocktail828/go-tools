package flatten

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrTypeConvert = errors.New("flatten: implicit type conversion is not allowed")
)

func flatten(prefix []string, mmp, result map[string]any) {
	for k, v := range mmp {
		if m, ok := v.(map[string]any); ok {
			flatten(append(prefix, k), m, result)
		} else {
			key := strings.Join(append(prefix, k), ".")
			result[key] = v
		}
	}
}

func setValue(fieldValue reflect.Value, value any) error {
	switch fieldValue.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			fieldValue.SetString(str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := value.(float64); ok { // JSON 数字默认是 float64
			fieldValue.SetInt(int64(num))
		} else {
			return errors.Wrapf(ErrTypeConvert, "cannot convert value to %s", fieldValue.Kind())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := value.(float64); ok { // JSON 数字默认是 float64
			fieldValue.SetUint(uint64(num))
		} else {
			return errors.Wrapf(ErrTypeConvert, "cannot convert value to %s", fieldValue.Kind())
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := value.(float64); ok { // JSON 数字默认是 float64
			fieldValue.SetFloat(num)
		} else {
			return errors.Wrapf(ErrTypeConvert, "cannot convert value to %s", fieldValue.Kind())
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			fieldValue.SetBool(b)
		} else {
			return errors.Wrapf(ErrTypeConvert, "cannot convert value to %s", fieldValue.Kind())
		}
	case reflect.Ptr:
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return setValue(fieldValue.Elem(), value)
	case reflect.Slice:
		if slice, ok := value.([]any); ok {
			elemType := fieldValue.Type().Elem()
			sliceValue := reflect.MakeSlice(fieldValue.Type(), len(slice), len(slice))
			for i, v := range slice {
				elemValue := reflect.New(elemType).Elem()
				if err := setValue(elemValue, v); err != nil {
					return err
				}
				sliceValue.Index(i).Set(elemValue)
			}
			fieldValue.Set(sliceValue)
		}
	case reflect.Map:
		if mp, ok := value.(map[string]any); ok {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
			}

			keyType := fieldValue.Type().Key()
			valueType := fieldValue.Type().Elem()

			for k, v := range mp {
				keyValue := reflect.New(keyType).Elem()
				if err := setValue(keyValue, k); err != nil {
					return errors.Wrapf(err, "fail to set map key: %s", k)
				}

				valueValue := reflect.New(valueType).Elem()
				if err := setValue(valueValue, v); err != nil {
					return errors.Wrapf(err, "fail to set map value for key: %s", k)
				}

				fieldValue.SetMapIndex(keyValue, valueValue)
			}
		} else {
			return errors.Errorf("unsupported map value type: %T", value)
		}
	default:
		return errors.Errorf("unsupported type: %s", fieldValue.Kind())
	}
	return nil
}

func lookup(mmp map[string]any, tag string) (any, bool) {
	var result any = mmp
	for _, key := range strings.Split(tag, ".") {
		if mp, ok := result.(map[string]any); ok {
			result = mp[key]
		} else {
			return nil, false
		}
	}

	return result, true
}

// parser json data into a flatten struct
func Unmarshal(data []byte, v any) error {
	mmp := map[string]any{}
	if err := json.Unmarshal(data, &mmp); err != nil {
		return err
	}

	flattened := map[string]any{}
	flatten(nil, mmp, flattened)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("v must point to a struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		tag := field.Tag.Get("flatten")
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue := rv.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		if fieldValue.Kind() == reflect.Map {
			if value, ok := lookup(mmp, tag); ok {
				if err := setValue(fieldValue, value); err != nil {
					return err
				}
			}
		} else {
			if value, ok := flattened[tag]; ok {
				if err := setValue(fieldValue, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type xCodec struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *xCodec {
	return &xCodec{r: r}
}

func (c *xCodec) Decode(v any) error {
	data, err := io.ReadAll(c.r)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}
