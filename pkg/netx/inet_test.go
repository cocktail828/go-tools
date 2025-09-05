package netx

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInet(t *testing.T) {
	_, err := Inet4()
	assert.NoError(t, err)

	_, err = Inet6()
	assert.NoError(t, err)
}

func TestTrans(t *testing.T) {
	ip := net.ParseIP("10.10.0.200")
	n, ok := IP2Long(ip)
	if !ok {
		t.Fatal("IP2Long failed")
	}
	assert.Equal(t, ip, Long2IP(n))
}
