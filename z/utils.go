package z

import (
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/z/locker"
	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/pkg/errors"
)

type random struct {
	sync.Mutex
	R *rand.Rand
}

var r = &random{
	R: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func GenerateRandomName() string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 8)

	locker.WithLock(r, func() {
		for i := range bytes {
			bytes[i] = chars[r.R.Intn(len(chars))]
		}
	})
	return string(bytes)
}

func Must(err error) {
	if !reflectx.IsNil(err) {
		panic(err)
	}
}

func BindEnv(name string, dst interface{}) (bool, error) {
	val := os.Getenv(name)
	if val == "" {
		return false, nil
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false, errors.New("dst should be a non-nil pointer")
	}

	rv = rv.Elem()
	if ok := rv.CanSet(); !ok {
		return false, errors.New("dst value cannot be changed")
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(val)

	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return false, err
		}
		rv.SetBool(v)

	case reflect.Float32:
		v, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return false, err
		}
		rv.SetFloat(v)

	case reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false, err
		}
		rv.SetFloat(v)

	case reflect.Int8:
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return false, err
		}
		rv.SetInt(v)

	case reflect.Int16:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return false, err
		}
		rv.SetInt(v)

	case reflect.Int32:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return false, err
		}
		rv.SetInt(v)

	case reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return false, err
		}
		rv.SetInt(v)

	case reflect.Uint8:
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return false, err
		}
		rv.SetUint(v)

	case reflect.Uint16:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return false, err
		}
		rv.SetUint(v)

	case reflect.Uint32:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return false, err
		}
		rv.SetUint(v)

	case reflect.Uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return false, err
		}
		rv.SetUint(v)

	default:
		return false, errors.Errorf("unsupport type: %v", rv.Kind())
	}
	return true, nil
}
