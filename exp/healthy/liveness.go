package healthy

import (
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Liveness interface {
	Probe() error
}

type SocketProbe struct {
	Addr    string
	Network string // tcp, udp
	Timeout time.Duration
}

func (p SocketProbe) Probe() error {
	c, err := net.DialTimeout(p.Network, p.Addr, p.Timeout)
	if err != nil {
		return err
	}

	c.Close()
	return nil
}

type HTTPProbe struct {
	URL     string
	Timeout time.Duration
}

func (p HTTPProbe) Probe() error {
	client := http.Client{Timeout: p.Timeout}
	resp, err := client.Get(p.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return errors.Errorf("expect 'http.StatusOK' but got '%d'", code)
	}
	return nil
}
