package httpx

import (
	"fmt"
	"io"
	"net/http"
)

type Unmarshaler func([]byte, interface{}) error

func Stringfy(b []byte, i interface{}) error {
	if s, ok := i.(*string); ok {
		*s = string(b)
		return nil
	}

	return fmt.Errorf("type assert fail: %T not *string", i)
}

type ResponseParser struct {
	unmarshaler Unmarshaler
}

func NewResponseParser(unmarshaler Unmarshaler) *ResponseParser {
	return &ResponseParser{unmarshaler: unmarshaler}
}

func (rp *ResponseParser) Parse(resp *http.Response, i interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return rp.ParseBody(body, i)
}

func (rp *ResponseParser) ParseBody(body []byte, i interface{}) error {
	return rp.unmarshaler(body, i)
}
