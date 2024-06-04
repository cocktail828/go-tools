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
