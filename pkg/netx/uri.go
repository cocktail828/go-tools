package netx

import (
	"net"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type ClusterURI struct {
	Scheme string      // protocol(http、mysql)
	User   string      // username(可选)
	Pass   string      // password(可选)
	Hosts  []HostEntry // host-def list(IP:PORT)
	Path   string      // Path
	Query  url.Values  // query params
}

func (c *ClusterURI) AllIPs() []string {
	var all []string
	for _, entry := range c.Hosts {
		all = append(all, entry.IPs...)
	}
	return all
}

type HostEntry struct {
	Raw string   // init host-def("api.example.com:80"、"10.0.0.1:8080")
	IPs []string // IP:PORT list
}

func ParseURI(rawURI string) (*ClusterURI, error) {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return nil, errors.Errorf("invalid uri format: %v", err)
	}

	if parsed.Scheme == "" {
		return nil, errors.New("missing scheme(ie.. http、mysql)")
	}

	var user, pass string
	if parsed.User != nil {
		user = parsed.User.Username()
		pass, _ = parsed.User.Password()
	}

	if parsed.Host == "" {
		return nil, errors.New("missing host-def")
	}

	hosts := []HostEntry{}
	for _, raw := range strings.Split(parsed.Host, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("host-def is empty")
		}

		host, port, err := net.SplitHostPort(raw)
		if err != nil {
			return nil, errors.Errorf("host-def format invalid(%s): %v", raw, err)
		}

		var ips []string
		if net.ParseIP(host) == nil {
			addrs, err := net.LookupHost(host)
			if err != nil {
				return nil, errors.Errorf("domain resolve failed(%s): %v", host, err)
			}
			for _, addr := range addrs {
				ips = append(ips, net.JoinHostPort(addr, port))
			}
		} else {
			// already ip address, just join host and port
			ips = []string{net.JoinHostPort(host, port)}
		}

		hosts = append(hosts, HostEntry{
			Raw: raw,
			IPs: ips,
		})
	}

	if len(hosts) == 0 {
		return nil, errors.New("host-def list is empty")
	}

	return &ClusterURI{
		Scheme: parsed.Scheme,
		User:   user,
		Pass:   pass,
		Hosts:  hosts,
		Path:   parsed.Path,
		Query:  parsed.Query(),
	}, nil
}
