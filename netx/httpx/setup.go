package httpx

import (
	"crypto/tls"
	"net/http"

	"github.com/cocktail828/go-tools/z/environs"
)

func init() {
	if v, _ := environs.Bool("SSL_NO_VERIFY"); v {
		InsecureSSL(v)
	}
}

func InsecureSSL(v bool) {
	transport := http.DefaultTransport.(*http.Transport)
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	transport.TLSClientConfig.InsecureSkipVerify = v
}
