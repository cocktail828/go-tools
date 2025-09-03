package httpx

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
)

func init() {
	envVal := os.Getenv("NO_SSL_VERIFY")
	if envVal == "" {
		return
	}

	insecure, err := strconv.ParseBool(envVal)
	if err != nil {
		return
	}
	InsecureSSL(insecure)
}

func InsecureSSL(v bool) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	transport.TLSClientConfig.InsecureSkipVerify = v
}
