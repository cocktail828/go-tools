package mdns

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/hashicorp/mdns"
)

type Service struct {
	Name     string   // required, service name, e.g. "_fnos"
	Service  string   // required, service, e.g. "_nas"
	Port     int      // required, service port, e.g. 8000
	Info     []string // optional, service info, e.g. "My awesome service"
	HostName string   // optional, host name, default is localhost
	Domain   string   // optional, domain name, default "local."
	IPs      []net.IP // optional, IP addresses for the service's address
}

func (s Service) Register(ctx context.Context) error {
	svc, err := mdns.NewMDNSService(s.Name, s.Service, s.Domain, s.HostName, s.Port, s.IPs, s.Info)
	if err != nil {
		return err
	}

	srv, err := mdns.NewServer(&mdns.Config{Zone: svc})
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

// Lookup queries for services of the given name and domain.
//
// If domain is empty, "local." is used.
func Lookup(ctx context.Context, service, domain string) ([]Entry, error) {
	entryCh := make(chan *mdns.ServiceEntry, 100)
	if err := mdns.QueryContext(ctx, &mdns.QueryParam{
		Service: service,
		Domain:  domain,
		Entries: entryCh,
		Logger:  log.New(io.Discard, "mdns: ", log.LstdFlags),
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
