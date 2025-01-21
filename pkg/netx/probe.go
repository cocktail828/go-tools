package netx

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
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

func ProbeHTTP(url string, tmo time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), tmo)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return errors.Errorf("expect 'http.StatusOK' but got '%d'", code)
	}
	return nil
}
