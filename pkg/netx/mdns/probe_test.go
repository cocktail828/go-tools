package mdns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestNextInstanceName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"My Service", "My Service (2)"},
		{"My Service (2)", "My Service (3)"},
		{"My Service (9)", "My Service (10)"},
		{"hostname", "hostname (2)"},
		// Parenthesised text that is not a numeric suffix is left intact and a
		// fresh suffix is appended.
		{"Service (beta)", "Service (beta) (2)"},
		// Empty parens are not a numeric suffix.
		{"Service ()", "Service () (2)"},
	}
	for _, c := range cases {
		if got := nextInstanceName(c.in); got != c.want {
			t.Errorf("nextInstanceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMsgConflictsWith(t *testing.T) {
	name := "orig._http._tcp.local."

	answerMsg := new(dns.Msg)
	answerMsg.Answer = append(answerMsg.Answer, &dns.SRV{
		Hdr:    dns.RR_Header{Name: name, Rrtype: dns.TypeSRV, Class: dns.ClassINET},
		Target: "other.local.",
		Port:   80,
	})
	if !msgConflictsWith(answerMsg, name) {
		t.Error("expected conflict for matching answer name")
	}
	// Matching should be case-insensitive per DNS name comparison rules.
	if !msgConflictsWith(answerMsg, "ORIG._http._tcp.local.") {
		t.Error("expected case-insensitive conflict match")
	}

	extraMsg := new(dns.Msg)
	extraMsg.Extra = append(extraMsg.Extra, &dns.A{
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET},
		A:   net.ParseIP("192.168.0.1"),
	})
	if !msgConflictsWith(extraMsg, name) {
		t.Error("expected conflict for matching name in Extra section")
	}

	otherMsg := new(dns.Msg)
	otherMsg.Answer = append(otherMsg.Answer, &dns.SRV{
		Hdr: dns.RR_Header{Name: "someoneelse._http._tcp.local.", Rrtype: dns.TypeSRV},
	})
	if msgConflictsWith(otherMsg, name) {
		t.Error("did not expect conflict for unrelated name")
	}

	if msgConflictsWith(new(dns.Msg), name) {
		t.Error("did not expect conflict for empty message")
	}
}

// fakeProber returns a conflict for any candidate name in the busy set, letting
// us drive the rename loop deterministically without real networking.
type fakeProber struct {
	busy     map[string]bool
	all      bool // when true, every name is reported as taken
	err      error
	probed   []string
	failFrom int // return err once len(probed) reaches this (0 = never)
}

func (f *fakeProber) probeName(_ context.Context, name string) (bool, error) {
	f.probed = append(f.probed, name)
	if f.failFrom > 0 && len(f.probed) >= f.failFrom {
		return false, f.err
	}
	if f.all {
		return true, nil
	}
	return f.busy[name], nil
}

func newService(t *testing.T, instance string) *MDNSService {
	t.Helper()
	m, err := NewMDNSService(instance, "_http._tcp", "local.", "testhost.", 80,
		[]net.IP{net.ParseIP("192.168.0.1")}, nil)
	if err != nil {
		t.Fatalf("NewMDNSService: %v", err)
	}
	return m
}

func TestProbeNoConflict(t *testing.T) {
	m := newService(t, "orig")
	fp := &fakeProber{busy: map[string]bool{}}

	if err := probe(context.Background(), m, fp, defaultMaxRenames); err != nil {
		t.Fatalf("probe: %v", err)
	}
	if m.Instance != "orig" {
		t.Errorf("Instance = %q, want unchanged %q", m.Instance, "orig")
	}
	if len(fp.probed) != 1 {
		t.Errorf("probed %d names, want 1: %v", len(fp.probed), fp.probed)
	}
}

func TestProbeRenamesOnConflict(t *testing.T) {
	m := newService(t, "orig")
	// The original and its first alternative are taken; the second is free.
	fp := &fakeProber{busy: map[string]bool{
		"orig._http._tcp.local.":     true,
		"orig (2)._http._tcp.local.": true,
		"orig (3)._http._tcp.local.": false,
	}}

	if err := probe(context.Background(), m, fp, defaultMaxRenames); err != nil {
		t.Fatalf("probe: %v", err)
	}
	if m.Instance != "orig (3)" {
		t.Errorf("Instance = %q, want %q", m.Instance, "orig (3)")
	}
	want := "orig (3)._http._tcp.local."
	if m.instanceAddr != want {
		t.Errorf("instanceAddr = %q, want %q", m.instanceAddr, want)
	}
	wantProbed := []string{
		"orig._http._tcp.local.",
		"orig (2)._http._tcp.local.",
		"orig (3)._http._tcp.local.",
	}
	if fmt.Sprint(fp.probed) != fmt.Sprint(wantProbed) {
		t.Errorf("probed sequence = %v, want %v", fp.probed, wantProbed)
	}
}

func TestProbeExhaustsRenames(t *testing.T) {
	m := newService(t, "orig")
	// Every name is taken; probing must give up after maxRenames renames.
	fp := &fakeProber{all: true}

	const maxRenames = 3
	err := probe(context.Background(), m, fp, maxRenames)
	if err == nil {
		t.Fatal("expected error when all names conflict, got nil")
	}
	// 1 initial probe + maxRenames additional probes.
	if len(fp.probed) != maxRenames+1 {
		t.Errorf("probed %d names, want %d: %v", len(fp.probed), maxRenames+1, fp.probed)
	}
}

func TestProbePropagatesError(t *testing.T) {
	m := newService(t, "orig")
	sentinel := fmt.Errorf("probe failed")
	fp := &fakeProber{busy: map[string]bool{}, err: sentinel, failFrom: 1}

	err := probe(context.Background(), m, fp, defaultMaxRenames)
	if err != sentinel {
		t.Fatalf("probe error = %v, want %v", err, sentinel)
	}
}

func TestProbeContextCancel(t *testing.T) {
	m := newService(t, "orig")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fp := &ctxProber{}
	err := probe(ctx, m, fp, defaultMaxRenames)
	if err != context.Canceled {
		t.Fatalf("probe error = %v, want %v", err, context.Canceled)
	}
}

// ctxProber always reports the context error, simulating a prober that honours
// cancellation.
type ctxProber struct{}

func (c *ctxProber) probeName(ctx context.Context, _ string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// bareZone implements Zone but not ProbeableZone, to check that NewServer
// rejects Config.Probe against a zone it cannot rename.
type bareZone struct{}

func (bareZone) Records(dns.Question) []dns.RR { return nil }

func TestNewServerProbeRequiresProbeableZone(t *testing.T) {
	_, err := NewServer(&Config{Zone: bareZone{}, Probe: &ProbeConfig{}})
	if err == nil {
		t.Fatal("expected error when Config.Probe is set on a non-ProbeableZone, got nil")
	}
	if !strings.Contains(err.Error(), "ProbeableZone") {
		t.Errorf("error = %q, want it to mention ProbeableZone", err)
	}
}

// compile-time assertion that *MDNSService satisfies ProbeableZone.
var _ ProbeableZone = (*MDNSService)(nil)

// TestProbeIntegrationRenames exercises the full multicast probing path: a
// server advertises "conflict" on the loopback interface, and a second service
// probing for the same name must detect the conflict and rename itself.
//
// This relies on working multicast loopback, which is unavailable in some
// sandboxed CI environments; it is skipped under -short and tolerates setup
// failures rather than reporting a spurious test failure.
func TestProbeIntegrationRenames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}

	iface := loopbackMulticastIface(t)

	existing := newService(t, "conflict")
	server, err := NewServer(&Config{Zone: existing, Iface: iface})
	if err != nil {
		t.Skipf("cannot start mdns server (multicast unavailable?): %v", err)
	}
	defer server.Shutdown()

	// Give the responder a moment to start listening.
	time.Sleep(100 * time.Millisecond)

	svc := newService(t, "conflict")
	cfg := &ProbeConfig{
		Attempts:   2,
		Interval:   200 * time.Millisecond,
		MaxRenames: 3,
		Iface:      iface,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Probe(ctx, svc, cfg); err != nil {
		t.Fatalf("Probe: %v", err)
	}

	if svc.Instance == "conflict" {
		t.Errorf("expected instance to be renamed away from %q, but it was unchanged", "conflict")
	}
	if svc.Instance != "conflict (2)" {
		t.Logf("instance renamed to %q (expected %q, but any non-conflicting name is acceptable)",
			svc.Instance, "conflict (2)")
	}
}

// TestProbeIntegrationNoConflict verifies that probing for a name nobody
// advertises leaves the instance name unchanged.
func TestProbeIntegrationNoConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multicast integration test in short mode")
	}

	iface := loopbackMulticastIface(t)

	svc := newService(t, "unique-name-nobody-has")
	cfg := &ProbeConfig{
		Attempts:   2,
		Interval:   200 * time.Millisecond,
		MaxRenames: 3,
		Iface:      iface,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Probe(ctx, svc, cfg); err != nil {
		t.Skipf("Probe failed (multicast unavailable?): %v", err)
	}

	if svc.Instance != "unique-name-nobody-has" {
		t.Errorf("instance = %q, want unchanged %q", svc.Instance, "unique-name-nobody-has")
	}
}

// loopbackMulticastIface returns an interface suitable for multicast testing.
// A loopback interface that supports multicast is preferred (it keeps traffic
// local); otherwise any up multicast-capable interface is used. The test is
// skipped if none is available.
func loopbackMulticastIface(t *testing.T) *net.Interface {
	t.Helper()
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Skipf("cannot list interfaces: %v", err)
	}
	var fallback *net.Interface
	for i := range ifaces {
		f := ifaces[i].Flags
		if f&net.FlagUp == 0 || f&net.FlagMulticast == 0 {
			continue
		}
		if f&net.FlagLoopback != 0 {
			return &ifaces[i]
		}
		if fallback == nil {
			fallback = &ifaces[i]
		}
	}
	if fallback != nil {
		return fallback
	}
	t.Skip("no multicast-capable interface available")
	return nil
}
