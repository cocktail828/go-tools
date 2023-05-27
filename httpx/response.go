package httpx

import (
	"fmt"
)

type Unmarshaler func([]byte, interface{}) error

func Stringfy(b []byte, i interface{}) error {
	if s, ok := i.(*string); ok {
		*s = string(b)
		return nil
	}

	return fmt.Errorf("type assert fail: %T not *string", i)
}

func ParseBody[T interface{}](body []byte, unmarshaler Unmarshaler) (t T, err error) {
	err = unmarshaler(body, &t)
	return
}
