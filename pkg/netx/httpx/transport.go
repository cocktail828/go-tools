package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cocktail828/go-tools/algo/balancer"
	"github.com/cocktail828/go-tools/exp/healthy"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"golang.org/x/sync/singleflight"
)

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

// NewTransport creates a new http.RoundTripper with the given insecureSkipVerify.
// and the other parameters is the same as http.DefaultTransport.
func NewTransport(insecureSkipVerify bool) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}
}

type inEndpoint struct {
	Addr string
}

func (e inEndpoint) MarkFailure()  {}
func (e inEndpoint) Healthy() bool { return true }
func (e inEndpoint) Weight() int   { return 100 }
func (e inEndpoint) Value() any    { return e.Addr }

// lookupInternal performs a DNS lookup for the given hostport.
// It returns a slice of healthy nodes for the given hostport.
//
// Parameters:
// - hostport: the hostport to lookup, e.g. "example.com:80".
//
// Returns:
// - []balancer.Node: the healthy nodes for the given hostport.
//
// If the hostport is not in the format of "host:port", it will return nil.
// If the host lookup fails, it will return nil.
// If no healthy node is found, it will return nil.
//
// Otherwise, it will return a slice of inEndpoint, each inEndpoint represents a healthy node.
func lookupInternal(hostport string) []balancer.Node {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil
	}

	hosts, err := net.LookupHost(host)
	if err != nil {
		return nil
	}

	healthyNodes := make([]balancer.Node, 0, len(hosts))
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip == nil || ip.To4() == nil {
			continue
		}

		addr := net.JoinHostPort(host, port)
		probe := healthy.SocketProbe{Addr: addr, Network: "tcp", Timeout: time.Millisecond * 100}
		if err := probe.Probe(); err == nil {
			healthyNodes = append(healthyNodes, inEndpoint{Addr: addr})
		} else {
			colorful.Warnf("probe target[%s] addr[%s] err: [%v]", hostport, addr, err)
		}
	}

	return healthyNodes
}

type lbRoundTripper struct {
	transport   http.Transport // the underlying round tripper to use
	builder     func(nodes []balancer.Node) balancer.Balancer
	portMap     map[string]string // http: 80, https: 443, ...
	lastCheckAt time.Time
	sg          singleflight.Group
	mu          sync.RWMutex
	selectorMap map[string]balancer.Balancer
}

func appendOnNonExist(portmap map[string]string, scheme string, port string) {
	if _, ok := portmap[scheme]; !ok {
		portmap[scheme] = port
	}
}

// NewDNSTransport creates a round tripper that uses DNS to lookup healthy nodes.
//
// Parameters:
// - transport: the underlying round tripper to use.
// - builder: the function to create a balancer for each hostport.
// - portmap: the map to map scheme to port, e.g. http: 80, https: 443, ...
//
// The round tripper will lookup healthy nodes on first request, and cache them for 3 seconds.
// It will use the user specified balancer to pick a node, if the balancer is not found,
// it will lookup healthy nodes and create a new balancer.
func NewDNSTransport(
	transport http.RoundTripper,
	builder func(nodes []balancer.Node) balancer.Balancer,
	portmap map[string]string,
) http.RoundTripper {
	if portmap == nil {
		portmap = map[string]string{}
	}

	appendOnNonExist(portmap, "http", "80")
	appendOnNonExist(portmap, "https", "443")

	return &lbRoundTripper{
		builder:     builder,
		portMap:     portmap,
		selectorMap: map[string]balancer.Balancer{},
	}
}

// create selector if not exists
func (lb *lbRoundTripper) createOnNonExist(hostport string) balancer.Balancer {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	selector, ok := lb.selectorMap[hostport]
	if !ok {
		healthyNodes := lookupInternal(hostport)
		selector = lb.builder(healthyNodes)
		lb.selectorMap[hostport] = selector
		lb.lastCheckAt = time.Now()
		return selector
	}
	return selector
}

func (lb *lbRoundTripper) Pick(hostport string) string {
	lb.mu.RLock()
	selector, ok := lb.selectorMap[hostport]
	lb.mu.RUnlock()
	if !ok {
		selector = lb.createOnNonExist(hostport)
	}

	if now := time.Now(); now.Sub(lb.lastCheckAt) > time.Second*3 {
		lb.sg.Do(hostport, func() (any, error) {
			lb.lastCheckAt = now
			healthyNodes := lookupInternal(hostport)
			selector.Update(healthyNodes)
			return "", nil
		})
	}

	val := selector.Pick()
	if val == nil {
		return ""
	}
	return val.Value().(string)
}

func (lb *lbRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	hostport := req.URL.Host
	if !strings.Contains(hostport, ":") {
		port, ok := lb.portMap[req.URL.Scheme]
		if !ok {
			// unknow port, fallback to default mode
			return lb.transport.RoundTrip(req)
		}

		hostport += ":" + port
	}
	old := req.URL.Host
	defer func() { req.URL.Host = old }()

	if addr := lb.Pick(hostport); addr != "" {
		req.URL.Host = addr
	}

	return lb.transport.RoundTrip(req)
}
