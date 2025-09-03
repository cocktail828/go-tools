package httpx

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/pkg/errors"
)

type Response struct {
	*http.Response
	loaded  atomic.Bool
	payload []byte
}

func (r *Response) prepare() error {
	if r.loaded.CompareAndSwap(false, true) {
		defer r.Response.Body.Close()
		payload, err := io.ReadAll(r.Response.Body)
		if err != nil {
			return errors.Errorf("read response payload fail: %v", err)
		}
		r.payload = payload
	}
	return nil
}

// Bind auto choose Content-Type and bind response body to v
func (r *Response) Bind(v any) error {
	if err := r.prepare(); err != nil {
		return err
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return errors.Errorf("Content-Type header is not set")
	}

	for _, ct := range strings.Split(strings.TrimSpace(contentType), ";") {
		switch {
		case strings.Contains(ct, "application/json"):
			return r.BindJSON(v)
		case strings.Contains(ct, "application/xml"), strings.Contains(ct, "text/xml"):
			return r.BindXML(v)
		default:
		}
	}
	return errors.Errorf("unsupported Content-Type: %s", contentType)
}

func (r *Response) BindJSON(v any) error {
	if err := r.prepare(); err != nil {
		return err
	}
	return json.Unmarshal(r.payload, v)
}

func (r *Response) BindXML(v any) error {
	if err := r.prepare(); err != nil {
		return err
	}
	return xml.Unmarshal(r.payload, v)
}
