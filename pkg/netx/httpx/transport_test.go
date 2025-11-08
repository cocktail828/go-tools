package httpx

import (
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/balancer"
)

func TestMain(m *testing.M) {
	// m.Run()
}

func TestDNSInsureTransport(t *testing.T) {
	c := http.Client{
		Transport: NewDNSTransport(
			http.DefaultTransport,
			func(nodes []balancer.Node) balancer.Balancer { return balancer.NewRoundRobin(nodes) },
			nil,
		),
		Timeout: time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://www.qq.com", nil)
	for range 10 {
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
}
