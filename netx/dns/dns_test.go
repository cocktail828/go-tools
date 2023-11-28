package dns_test

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/dns"
	"github.com/stretchr/testify/assert"
)

func TestProbe(t *testing.T) {
	assert.Equal(t, nil, dns.ProbeTCP("110.242.68.66:80", time.Millisecond*100))
	assert.Equal(t, nil, dns.ProbeUDP("8.8.8.8:53", time.Millisecond*100))
}

type TypicalErr struct {
	e string
}

func (t TypicalErr) Error() string {
	return t.e
}

func TestDNS(t *testing.T) {
	err := TypicalErr{"typical error"}
	err1 := fmt.Errorf("wrap err: %w", err)
	err2 := fmt.Errorf("wrap err1: %w", err1)
	var e TypicalErr
	if !errors.As(err2, &e) {
		panic("TypicalErr is not on the chain of err2")
	}

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
		fmt.Println(err.(*net.DNSError).Err == "no such host")
	}()
	func() {
		rr, err := r.LookupIP("baidu.com")
		assert.Equal(t, nil, err)
		if err == nil {
			fmt.Println(rr.ToIP(), rr.Equal(rr))
		}
	}()
}
