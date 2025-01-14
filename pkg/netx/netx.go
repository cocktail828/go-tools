package netx

import (
	"encoding/binary"
	"net"
)

// Long2IP converts an integer to a net.IP
func Long2IP(ipInt uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipInt)
	return ip
}

// IP2Long converts a net.IP to an integer
func IP2Long(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}
