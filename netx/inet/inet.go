package inet

import (
	"net"
	"strings"
)

func Inet(validater ...func(*net.IPNet) bool) ([]net.Addr, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var addrs []net.Addr
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			iaddrs, err := inter.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range iaddrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if func() bool {
						for _, f := range validater {
							if !f(ipnet) {
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
	return Inet(func(i *net.IPNet) bool {
		return !i.IP.IsLoopback()
	}, func(i *net.IPNet) bool {
		return i.IP.To4() != nil
	})
}

func Inet6() ([]net.Addr, error) {
	return Inet(func(i *net.IPNet) bool {
		return !i.IP.IsLoopback()
	}, func(i *net.IPNet) bool {
		return i.IP.To16() != nil && i.IP.To4() == nil
	})
}
