package mdns

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/mdns"
)

type Instance struct {
	Iface    *net.Interface // optional, binds the multicast listener to the given interface
	Name     string         // required, service name, e.g. "_fnos"
	Service  string         // required, service, e.g. "_nas"
	Domain   string         // optional, domain name, default "local."
	HostName string         // optional, host name, default is localhost
	Port     int            // required, service port, e.g. 8000
	IPs      []net.IP       // optional, IP addresses for the service's address
	Info     []string       // optional, service info, e.g. "My awesome service"
}

func Announce(ctx context.Context, inst Instance) error {
	svc, err := mdns.NewMDNSService(inst.Name, inst.Service, inst.Domain, inst.HostName, inst.Port, inst.IPs, inst.Info)
	if err != nil {
		return err
	}

	srv, err := mdns.NewServer(&mdns.Config{
		Zone:   svc,
		Iface:  inst.Iface,
		Logger: log.New(io.Discard, "mdns.query: ", log.LstdFlags),
	})
	if err != nil {
		return err
	}
	defer srv.Shutdown()

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

type Entry struct {
	Name   string
	Host   string
	AddrV4 net.IP
	AddrV6 *net.IPAddr
	Port   int
	Info   []string
}

type LookParam struct {
	Service             string         // Service to lookup
	Domain              string         // Lookup domain, default "local"
	Timeout             time.Duration  // Lookup timeout, default 1 second
	Interface           *net.Interface // Multicast interface to use
	WantUnicastResponse bool           // Unicast response desired, as per 5.4 in RFC
	DisableIPv4         bool           // Whether to disable usage of IPv4 for MDNS operations.
	DisableIPv6         bool           // Whether to disable usage of IPv6 for MDNS operations.
}

// Lookup queries for services of the given name and domain.
//
// If domain is empty, "local." is used.
func Lookup(p LookParam) ([]Entry, error) {
	entryCh := make(chan *mdns.ServiceEntry, 100)
	if err := mdns.Query(&mdns.QueryParam{
		Service:             p.Service,
		Domain:              p.Domain,
		Timeout:             p.Timeout,
		Interface:           p.Interface,
		Entries:             entryCh,
		WantUnicastResponse: p.WantUnicastResponse,
		DisableIPv4:         p.DisableIPv4,
		DisableIPv6:         p.DisableIPv6,
		Logger:              log.New(io.Discard, "mdns.response: ", log.LstdFlags),
	}); err != nil {
		return nil, err
	}
	close(entryCh)

	var instances []Entry
	for entry := range entryCh {
		instances = append(instances, Entry{
			Name:   entry.Name,
			Host:   entry.Host,
			AddrV4: entry.AddrV4,
			AddrV6: entry.AddrV6IPAddr,
			Port:   entry.Port,
			Info:   entry.InfoFields,
		})
	}
	return instances, nil
}
