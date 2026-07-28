package healthy

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSocketProbe(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer ln.Close()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	p := SocketProbe{Addr: ln.Addr().String(), Network: "tcp", Timeout: time.Second}
	assert.NoError(t, p.Probe())

	// A closed/unreachable address must yield an error.
	bad := SocketProbe{Addr: "127.0.0.1:1", Network: "tcp", Timeout: 200 * time.Millisecond}
	assert.Error(t, bad.Probe())
}

func TestHTTPProbe(t *testing.T) {
	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okSrv.Close()

	assert.NoError(t, HTTPProbe{URL: okSrv.URL, Timeout: time.Second}.Probe())

	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failSrv.Close()

	assert.Error(t, HTTPProbe{URL: failSrv.URL, Timeout: time.Second}.Probe())

	// An unreachable URL must yield a transport error.
	assert.Error(t, HTTPProbe{URL: "http://127.0.0.1:1", Timeout: 200 * time.Millisecond}.Probe())
}

func TestHealthChecker(t *testing.T) {
	rec := httptest.NewRecorder()
	HealthChecker{Ready: true, Message: "ok"}.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, true, body["ready"])
	assert.Equal(t, "ok", body["message"])

	// Not ready -> 403 for k8s readiness semantics.
	rec = httptest.NewRecorder()
	HealthChecker{Ready: false}.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
