package httpx

import (
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/algo/balancer"
)

func TestInsureTransport(t *testing.T) {
	uri := "https://aiportal.h3c.com:40212/snappyservice/profile/upload/ZJSZTB/virtualHuman.png"
	for _, b := range []bool{true, false} {
		cli := http.Client{Transport: NewTransport(b)}
		resp, err := cli.Get(uri)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
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

	req, _ := http.NewRequest(http.MethodGet, "http://evo-dx.xf-yun.com", nil)
	for range 10 {
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
}
