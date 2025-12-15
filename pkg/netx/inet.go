package netx

import (
	"encoding/binary"
	"net"
)

// Long2IP converts an integer to a net.IP
func Long2IP(v uint32) net.IP {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return net.IPv4(bs[0], bs[1], bs[2], bs[3])
}

// IP2Long converts a net.IP to an integer
// Returns 0 and false if the IP is not a valid IPv4 address
func IP2Long(ip net.IP) (uint32, bool) {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0, false
	}
	return binary.BigEndian.Uint32(ipv4), true
}

// the IP will be filter out if the result is true
type Excluder func(*net.IPNet) bool

func Inet(filters ...Excluder) ([]*net.IPNet, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var addrs []*net.IPNet
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
						addrs = append(addrs, ipnet)
					}
				}
			}
		}
	}

	return addrs, nil
}

func Inet4() ([]*net.IPNet, error) {
	return Inet(
		func(i *net.IPNet) bool { return i.IP.IsLoopback() },
		func(i *net.IPNet) bool { return i.IP.To4() == nil },
	)
}

func Inet6() ([]*net.IPNet, error) {
	return Inet(
		func(i *net.IPNet) bool { return i.IP.IsLoopback() },
		func(i *net.IPNet) bool { return i.IP.To16() == nil || i.IP.To4() != nil },
	)
}
