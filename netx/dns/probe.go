package dns

import (
	"net"
	"time"
)

type Prober func(addr string, tmo time.Duration) error

func probe(network string, addr string, tmo time.Duration) error {
	if c, err := net.DialTimeout(network, addr, tmo); err != nil {
		return err
	} else {
		c.Close()
		return nil
	}
}

func ProbeTCP(addr string, tmo time.Duration) error {
	return probe("tcp", addr, tmo)
}

func ProbeUDP(addr string, tmo time.Duration) error {
	return probe("udp", addr, tmo)
}
