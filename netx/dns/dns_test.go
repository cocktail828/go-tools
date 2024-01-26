package dns_test

import (
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/dns"
	"github.com/stretchr/testify/assert"
)

func TestProbe(t *testing.T) {
	assert.Equal(t, nil, dns.ProbeTCP("110.242.68.66:80", time.Millisecond*100))
	assert.Equal(t, nil, dns.ProbeUDP("8.8.8.8:53", time.Millisecond*100))
}

var (
	r     = dns.Resolver{}
	cases = map[string]struct {
		domain   string
		function func(host string) (dns.IPSet, error)
		pass     bool
	}{
		"LookupA":    {"alipay.com", r.LookupA, true},
		"LookupAAAA": {"alipay.com", r.LookupAAAA, true},
		"LookupIP":   {"alipay.com", r.LookupIP, true},
	}
)

func TestDNS(t *testing.T) {
	for name, val := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := val.function(val.domain)
			if val.pass != (err == nil) {
				t.Errorf("oops! expect pass but get error %v", err)
			}
		})
	}
}
