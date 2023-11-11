package dns_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/dns"
	"github.com/stretchr/testify/assert"
)

func TestProbe(t *testing.T) {
	assert.Equal(t, nil, dns.ProbeTCP("110.242.68.66:80", time.Millisecond*100))
	assert.Equal(t, nil, dns.ProbeUDP("8.8.8.8:53", time.Millisecond*100))
}

func TestDNS(t *testing.T) {
	r := dns.Resolver{}
	func() {
		rr, err := r.LookupA("baidu.com")
		assert.Equal(t, nil, err)
		if err == nil {
			fmt.Println(rr.ToIP(), rr.Equal(rr))
		}
	}()
	func() {
		rr, err := r.LookupAAAA("baidu.com")
		assert.Equal(t, nil, err)
		if err == nil {
			fmt.Println(rr.ToIP(), rr.Equal(rr))
		}
	}()
	func() {
		rr, err := r.LookupIP("baidu.com")
		assert.Equal(t, nil, err)
		if err == nil {
			fmt.Println(rr.ToIP(), rr.Equal(rr))
		}
	}()
}
