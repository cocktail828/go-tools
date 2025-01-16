package netx

import (
	"net"
)

// the IP will be filter out if the result is true
type Filter func(*net.IPNet) bool

func Inet(filters ...Filter) ([]net.Addr, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var addrs []net.Addr
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 {
			iaddrs, err := inter.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range iaddrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if func() bool {
						for _, f := range filters {
							if f(ipnet) {
								return false
							}
						}
						return true
					}() {
						addrs = append(addrs, addr)
					}
				}
			}
		}
	}

	return addrs, nil
}

func Inet4() ([]net.Addr, error) {
	return Inet(
		func(i *net.IPNet) bool { return i.IP.IsLoopback() },
		func(i *net.IPNet) bool { return i.IP.To4() == nil },
	)
}

func Inet6() ([]net.Addr, error) {
	return Inet(
		func(i *net.IPNet) bool { return i.IP.IsLoopback() },
		func(i *net.IPNet) bool { return i.IP.To16() == nil || i.IP.To4() != nil },
	)
}
