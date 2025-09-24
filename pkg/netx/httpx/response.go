package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type Response struct {
	*http.Response
	mu     sync.Mutex
	buffer *[]byte
}

func (r *Response) shouldLoad() error {
	if r.buffer != nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.buffer != nil {
		return nil
	}

	buf, err := io.ReadAll(r.Response.Body)
	if err != nil {
		nopBody := []byte{}
		r.buffer = &nopBody
		return errors.Errorf("read response payload fail: %v", err)
	}
	r.buffer = &buf
	return nil
}

func (r *Response) Payload() io.ReadCloser {
	if r.buffer == nil {
		r.shouldLoad()
	}
	return io.NopCloser(bytes.NewReader(*r.buffer))
}

// Bind auto choose Content-Type and bind response body to v
func (r *Response) Bind(v any) error {
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
	if err := r.shouldLoad(); err != nil {
		return err
	}
	return json.Unmarshal(*r.buffer, v)
}

func (r *Response) BindXML(v any) error {
	if err := r.shouldLoad(); err != nil {
		return err
	}
	return xml.Unmarshal(*r.buffer, v)
}
