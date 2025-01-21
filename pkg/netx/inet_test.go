package netx

import (
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInet(t *testing.T) {
	log.Println(Inet4())
	log.Println(Inet6())
}

func TestTrans(t *testing.T) {
	assert.Equal(t, net.IPv4(10, 10, 0, 200).String(), Long2IP(IP2Long(net.IPv4(10, 10, 0, 200))).String())
}
