package mdns

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// These are end-to-end tests: they stand up a real mDNS responder and use the
// package's own querier to discover it over multicast. They rely on working
// multicast loopback, which is unavailable in some sandboxed CI environments,
// so they skip under -short and treat server setup failure as a skip rather
// than a spurious failure.

// testService builds an MDNSService advertising a single IPv4 address, so
// discovery yields deterministic A records.
func testService(t *testing.T, instance, service string, txt []string) *MDNSService {
	t.Helper()
	m, err := NewMDNSService(instance, service, "local.", "testhost.", 8080,
		[]net.IP{net.ParseIP("192.168.0.1")}, txt)
	if err != nil {
		t.Fatalf("NewMDNSService: %v", err)
	}
	return m
}

// startServer starts an mDNS responder, skipping the test if multicast is
// unavailable. The server is shut down automatically at the end of the test.
func startServer(t *testing.T, cfg *Config) *Server {
	t.Helper()
	s, err := NewServer(cfg)
	if err != nil {
		t.Skipf("cannot start mdns server (multicast unavailable?): %v", err)
	}
	t.Cleanup(func() { s.Shutdown() })
	return s
}

// discover runs a service-discovery query and returns the collected entries.
func discover(t *testing.T, service string, iface *net.Interface, timeout time.Duration) []*ServiceEntry {
	t.Helper()

	entriesCh := make(chan *ServiceEntry, 16)
	var (
		mu  sync.Mutex
		got []*ServiceEntry
	)
	done := make(chan struct{})
	go func() {
		for e := range entriesCh {
			mu.Lock()
			got = append(got, e)
			mu.Unlock()
		}
		close(done)
	}()

	params := &QueryParam{
		Service:   service,
		Domain:    "local",
		Timeout:   timeout,
		Interface: iface,
		Entries:   entriesCh,
	}
	err := Query(params)
	close(entriesCh)
	<-done
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	return got
}

// findEntry returns the discovered entry whose name matches instanceAddr, or
// nil if it was not found. Discovered names arrive in DNS presentation format,
// where characters such as spaces and parentheses (present in a renamed
// instance label like "dup (2)") are backslash-escaped, so both sides are
// unescaped before comparison.
func findEntry(entries []*ServiceEntry, instanceAddr string) *ServiceEntry {
	want := unescapeDNSName(instanceAddr)
	for _, e := range entries {
		if strings.EqualFold(unescapeDNSName(e.Name), want) {
			return e
		}
	}
	return nil
}

// unescapeDNSName decodes DNS presentation-format escaping (RFC 1035 §5.1):
// "\DDD" (three decimal digits) becomes the byte with that value, and "\c"
// (any other escaped character) becomes the literal character.
func unescapeDNSName(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		// Escaped sequence.
		if i+3 < len(s) && isDigit(s[i+1]) && isDigit(s[i+2]) && isDigit(s[i+3]) {
			n := int(s[i+1]-'0')*100 + int(s[i+2]-'0')*10 + int(s[i+3]-'0')
			b.WriteByte(byte(n))
			i += 3
			continue
		}
		b.WriteByte(s[i+1])
		i++
	}
	return b.String()
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// TestRegisterAndDiscover registers a service without probing and verifies that
// a query discovers it with the expected host, port, address and TXT info.
func TestRegisterAndDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}
	iface := loopbackMulticastIface(t)

	svc := testService(t, "reg-plain", "_http._tcp", []string{"path=/health"})
	startServer(t, &Config{Zone: svc, Iface: iface})

	// Give the responder a moment to start listening.
	time.Sleep(100 * time.Millisecond)

	entries := discover(t, "_http._tcp", iface, 2*time.Second)
	e := findEntry(entries, svc.instanceAddr)
	if e == nil {
		t.Fatalf("service %q not discovered; got %d entries: %v", svc.instanceAddr, len(entries), entries)
	}

	if e.Port != 8080 {
		t.Errorf("Port = %d, want 8080", e.Port)
	}
	if e.Host != "testhost." {
		t.Errorf("Host = %q, want %q", e.Host, "testhost.")
	}
	if e.AddrV4 == nil || !e.AddrV4.Equal(net.ParseIP("192.168.0.1")) {
		t.Errorf("AddrV4 = %v, want 192.168.0.1", e.AddrV4)
	}
	if strings.Join(e.TXT, "|") != "path=/health" {
		t.Errorf("Info = %q, want %q", e.TXT[0], "path=/health")
	}
}

// TestRegisterWithProbeAndDiscover registers a service with probing enabled
// (no conflict present) and verifies the name is unchanged and discoverable.
func TestRegisterWithProbeAndDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}
	iface := loopbackMulticastIface(t)

	svc := testService(t, "reg-probed", "_http._tcp", nil)
	startServer(t, &Config{
		Zone:  svc,
		Iface: iface,
		Probe: &ProbeConfig{
			Attempts:    2,
			Interval:    200 * time.Millisecond,
			MaxRenames:  3,
			Iface:       iface,
			DisableIPv6: true,
		},
	})

	// No conflicting responder exists, so the name must be preserved.
	if svc.Instance != "reg-probed" {
		t.Errorf("Instance = %q, want unchanged %q", svc.Instance, "reg-probed")
	}

	time.Sleep(100 * time.Millisecond)

	entries := discover(t, "_http._tcp", iface, 2*time.Second)
	if findEntry(entries, svc.instanceAddr) == nil {
		t.Fatalf("probed service %q not discovered; got %d entries", svc.instanceAddr, len(entries))
	}
}

// TestRegisterWithProbeRenamesOnConflict registers one service, then registers
// a second with the same instance name and probing enabled. The second must
// rename itself, and both distinct names must then be discoverable.
func TestRegisterWithProbeRenamesOnConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}
	iface := loopbackMulticastIface(t)

	first := testService(t, "dup", "_http._tcp", nil)
	startServer(t, &Config{Zone: first, Iface: iface})
	time.Sleep(100 * time.Millisecond)

	second := testService(t, "dup", "_http._tcp", nil)
	startServer(t, &Config{
		Zone:  second,
		Iface: iface,
		Probe: &ProbeConfig{
			Attempts:    2,
			Interval:    250 * time.Millisecond,
			MaxRenames:  5,
			Iface:       iface,
			DisableIPv6: true,
		},
	})

	if second.Instance == "dup" {
		t.Fatalf("second service kept conflicting name %q; expected a rename", second.Instance)
	}
	if first.Instance != "dup" {
		t.Errorf("first service name changed to %q, want %q", first.Instance, "dup")
	}

	time.Sleep(100 * time.Millisecond)

	entries := discover(t, "_http._tcp", iface, 3*time.Second)
	if findEntry(entries, first.instanceAddr) == nil {
		t.Errorf("first service %q not discovered", first.instanceAddr)
	}
	if findEntry(entries, second.instanceAddr) == nil {
		t.Errorf("renamed service %q not discovered", second.instanceAddr)
	}
}

// TestDiscoverNoService verifies a query for an unadvertised service returns no
// entries within the timeout.
func TestDiscoverNoService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}
	iface := loopbackMulticastIface(t)

	// Start a responder for a different service so the network is live but has
	// nothing matching our query.
	svc := testService(t, "other", "_other._tcp", nil)
	startServer(t, &Config{Zone: svc, Iface: iface})
	time.Sleep(100 * time.Millisecond)

	entries := discover(t, "_nonexistent._tcp", iface, 1*time.Second)
	if len(entries) != 0 {
		t.Errorf("expected no entries for unadvertised service, got %d: %v", len(entries), entries)
	}
}
