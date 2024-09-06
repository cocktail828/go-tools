package netx

import (
	"net"
	"time"
)

type Prober func(addr string, tmo time.Duration) error

func Probe(network string, addr string, tmo time.Duration) error {
	if c, err := net.DialTimeout(network, addr, tmo); err != nil {
		return err
	} else {
		c.Close()
		return nil
	}
}

func ProbeTCP(addr string, tmo time.Duration) error {
	return Probe("tcp", addr, tmo)
}

func ProbeUDP(addr string, tmo time.Duration) error {
	return Probe("udp", addr, tmo)
}
