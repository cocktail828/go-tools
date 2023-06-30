package dns

import (
	"context"
	"net"
)

func LookupSRV(service, proto, name string) ([]*net.SRV, error) {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

func LookupA(name string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(context.Background(), "ip4", name)
}

func LookupAAAA(name string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(context.Background(), "ip6", name)
}

func LookupPTR(addr string) (names []string, err error) {
	return net.LookupAddr(addr)
}
