// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !js && !windows

// Read system DNS config from /etc/resolv.conf

package net

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Parse IPv4 address (d.d.d.d).
func ParseIPv4(s string) net.IP {
	var p [net.IPv4len]byte
	for i := 0; i < net.IPv4len; i++ {
		if len(s) == 0 {
			// Missing octets.
			return nil
		}
		if i > 0 {
			if s[0] != '.' {
				return nil
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return nil
		}
		if c > 1 && s[0] == '0' {
			// Reject non-zero components with leading zeroes.
			return nil
		}
		s = s[c:]
		p[i] = byte(n)
	}
	if len(s) != 0 {
		return nil
	}
	return net.IPv4(p[0], p[1], p[2], p[3])
}

// ParseIPv6 parses s as a literal IPv6 address described in RFC 4291
// and RFC 5952.
func ParseIPv6(s string) (ip net.IP) {
	ip = make(net.IP, net.IPv6len)
	ellipsis := -1 // position of ellipsis in ip

	// Might have leading ellipsis
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// Might be only ellipsis
		if len(s) == 0 {
			return ip
		}
	}

	// Loop, parsing hex numbers followed by colon.
	i := 0
	for i < net.IPv6len {
		// Hex number.
		n, c, ok := xtoi(s)
		if !ok || n > 0xFFFF {
			return nil
		}

		// If followed by dot, might be in trailing IPv4.
		if c < len(s) && s[c] == '.' {
			if ellipsis < 0 && i != net.IPv6len-net.IPv4len {
				// Not the right place.
				return nil
			}
			if i+net.IPv4len > net.IPv6len {
				// Not enough room.
				return nil
			}
			ip4 := ParseIPv4(s)
			if ip4 == nil {
				return nil
			}
			ip[i] = ip4[12]
			ip[i+1] = ip4[13]
			ip[i+2] = ip4[14]
			ip[i+3] = ip4[15]
			s = ""
			i += net.IPv4len
			break
		}

		// Save this 16-bit chunk.
		ip[i] = byte(n >> 8)
		ip[i+1] = byte(n)
		i += 2

		// Stop at end of string.
		s = s[c:]
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' || len(s) == 1 {
			return nil
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return nil
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return nil
	}

	// If didn't parse enough, expand ellipsis.
	if i < net.IPv6len {
		if ellipsis < 0 {
			return nil
		}
		n := net.IPv6len - i
		for j := i - 1; j >= ellipsis; j-- {
			ip[j+n] = ip[j]
		}
		for j := ellipsis + n - 1; j >= ellipsis; j-- {
			ip[j] = 0
		}
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return nil
	}
	return ip
}

// ParseIP parses s as an IP address, returning the result.
// The string s can be in IPv4 dotted decimal ("192.0.2.1"), IPv6
// ("2001:db8::68"), or IPv4-mapped IPv6 ("::ffff:192.0.2.1") form.
// If s is not a valid textual representation of an IP address,
// ParseIP returns nil.
func ParseIP(s string) net.IP {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ParseIPv4(s)
		case ':':
			return ParseIPv6(s)
		}
	}
	return nil
}

// ParseIPZone parses s as an IP address, return it and its associated zone
// identifier (IPv6 only).
func ParseIPZone(s string) (net.IP, string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ParseIPv4(s), ""
		case ':':
			return ParseIPv6Zone(s)
		}
	}
	return nil, ""
}

// ParseIPv6Zone parses s as a literal IPv6 address and its associated zone
// identifier which is described in RFC 4007.
func ParseIPv6Zone(s string) (net.IP, string) {
	// The IPv6 scoped addressing zone identifier starts after the
	// last percent sign.
	if i := strings.LastIndexByte(s, '%'); i > 0 {
		return ParseIPv6(s[:i]), s[i+1:]
	} else {
		return ParseIPv6(s), ""
	}
}

// ipVersion returns the provided network's IP version: '4', '6' or 0
// if network does not end in a '4' or '6' byte.
func IpVersion(network string) byte {
	if network == "" {
		return 0
	}
	n := network[len(network)-1]
	if n != '4' && n != '6' {
		n = 0
	}
	return n
}

// Get outbound Ip addr
func OutboundAddr() (net.IP, error) {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", time.Millisecond*10)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}

type SRV struct {
	net.SRV
	IPs []net.IP
}

// lookup SRV, then lookup target
func LookupSRV(service, proto, name string) ([]SRV, error) {
	_, srvs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return nil, err
	}

	records := []SRV{}
	for _, srv := range srvs {
		records = append(records, SRV{
			SRV: *srv,
			IPs: func() []net.IP {
				ips, err := net.LookupIP(srv.Target)
				if err != nil {
					return nil
				}
				return ips
			}(),
		})
	}

	return records, nil
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

type SOA dns.SOA

func LookupSOA(name string) (*SOA, error) {
	conf := loadResolverConf()

	for _, n := range conf.NameList(name) {
		for i := 0; i < len(conf.Servers); i++ {
			server := conf.Servers[conf.serverOffset()]
			c := dns.Client{
				Net:            "udp",
				Timeout:        conf.Timeout,
				SingleInflight: conf.SingleRequest,
			}
			if conf.UseTCP {
				c.Net = "tcp"
			}

			m := dns.Msg{}
			m.AuthenticatedData = true
			m.SetQuestion(n, dns.TypeSOA)
			rmsg, _, err := c.Exchange(&m, server)
			if err != nil {
				continue
			}

			for _, answer := range rmsg.Answer {
				if soa, ok := answer.(*dns.SOA); ok {
					return (*SOA)(soa), nil
				}
			}
		}
	}

	return nil, errors.New("no such record found")
}
