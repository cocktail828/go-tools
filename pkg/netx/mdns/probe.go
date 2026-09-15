package mdns

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	// defaultProbeAttempts is the number of probe queries sent for a single
	// candidate name before it is considered free. RFC 6762 section 8.1
	// recommends sending three probes.
	defaultProbeAttempts = 3

	// defaultProbeInterval is the delay between probe queries, and the window
	// during which a conflicting response is awaited. RFC 6762 section 8.1
	// recommends 250ms.
	defaultProbeInterval = 250 * time.Millisecond

	// defaultMaxRenames caps how many alternative names are tried before
	// probing gives up. This bounds the work when a network is saturated with
	// conflicting responders.
	defaultMaxRenames = 10
)

// nameSuffixRe matches an instance name that already carries a numeric
// disambiguation suffix, e.g. "My Service (2)".
var nameSuffixRe = regexp.MustCompile(`^(.*) \((\d+)\)$`)

// nextInstanceName returns the next candidate instance name following the
// Bonjour convention recommended by RFC 6762 section 9: a numeric suffix in
// parentheses that is incremented on each conflict.
//
//	"My Service"     -> "My Service (2)"
//	"My Service (2)" -> "My Service (3)"
func nextInstanceName(name string) string {
	if m := nameSuffixRe.FindStringSubmatch(name); m != nil {
		// The captured group is a run of digits, so Atoi cannot fail in a way
		// that matters; on overflow we fall back to restarting the sequence.
		if n, err := strconv.Atoi(m[2]); err == nil {
			return fmt.Sprintf("%s (%d)", m[1], n+1)
		}
	}
	return name + " (2)"
}

// msgConflictsWith reports whether msg contains any record whose owner name
// matches name. During probing, a response bearing the name we intend to claim
// means another host already owns it (RFC 6762 section 8.1).
func msgConflictsWith(msg *dns.Msg, name string) bool {
	for _, rr := range append(append([]dns.RR{}, msg.Answer...), msg.Extra...) {
		if strings.EqualFold(rr.Header().Name, name) {
			return true
		}
	}
	return false
}

// prober sends probe queries for a candidate name and reports whether the name
// is already claimed on the network. It is an interface so the rename loop can
// be exercised without real multicast traffic.
type Prober interface {
	// probeName reports whether name is already in use. It returns an error
	// only when probing itself fails (for example, the context is cancelled).
	probeName(ctx context.Context, name string) (conflict bool, err error)
}

// ProbeableZone is a Zone whose unique record name can be probed for conflicts
// and, on conflict, changed to an alternative. Probe and NewServer's automatic
// probing operate purely through this interface, so they never depend on any
// concrete Zone type. *MDNSService implements it.
type ProbeableZone interface {
	Zone

	// ProbeName returns the fully-qualified owner name that must be unique on
	// the network (for MDNSService, the service instance address).
	ProbeName() string

	// Rename selects the next candidate identity after a conflict is detected,
	// so that a subsequent ProbeName reflects the new name.
	Rename()
}

// ProbeConfig tunes the probing performed by MDNSService.Probe. A nil config
// uses defaults compatible with RFC 6762.
type ProbeConfig struct {
	// Attempts is the number of probe queries per candidate name. Defaults to
	// defaultProbeAttempts.
	Attempts int

	// Interval is the delay between probe queries and the window during which a
	// conflicting response is awaited. Defaults to defaultProbeInterval.
	Interval time.Duration

	// MaxRenames caps how many alternative names are tried before giving up.
	// Defaults to defaultMaxRenames.
	MaxRenames int

	// Iface optionally binds probing to a specific multicast interface.
	Iface *net.Interface

	// DisableIPv4 and DisableIPv6 restrict the address families used for
	// probing. At least one family must remain enabled.
	DisableIPv4 bool
	DisableIPv6 bool
}

func (c *ProbeConfig) withDefaults() *ProbeConfig {
	out := ProbeConfig{}
	if c != nil {
		out = *c
	}
	if out.Attempts <= 0 {
		out.Attempts = defaultProbeAttempts
	}
	if out.Interval <= 0 {
		out.Interval = defaultProbeInterval
	}
	if out.MaxRenames <= 0 {
		out.MaxRenames = defaultMaxRenames
	}
	return &out
}

// Probe implements the conflict-detection and renaming phase of RFC 6762. It
// sends probe queries for the zone's unique name and, if another host is
// already advertising that name, transparently renames the zone (via
// ProbeableZone.Rename) and probes again until an unused name is found or
// MaxRenames is exhausted.
//
// Probe should be called before the zone is handed to NewServer, so that the
// mutation of the name does not race with the server's reads. NewServer can run
// this automatically; see Config.Probe.
func Probe(ctx context.Context, z ProbeableZone, cfg *ProbeConfig) error {
	cfg = cfg.withDefaults()

	p, err := newMulticastProber(cfg)
	if err != nil {
		return err
	}
	defer p.Close()

	return probe(ctx, z, p, cfg.MaxRenames)
}

// probe runs the rename loop against an arbitrary prober. It is the testable
// core of Probe and depends only on the ProbeableZone interface.
func probe(ctx context.Context, z ProbeableZone, p Prober, maxRenames int) error {
	for renames := 0; ; renames++ {
		name := z.ProbeName()
		conflict, err := p.probeName(ctx, name)
		if err != nil {
			return err
		}
		if !conflict {
			// The name is free; it is ours to claim.
			return nil
		}
		if renames >= maxRenames {
			return fmt.Errorf("mdns: no available name for %q after %d renames", name, maxRenames)
		}
		z.Rename()
	}
}

// multicastProber is the network-backed prober used by Probe. It reuses the
// query client to send probes and listen for conflicting responses.
type multicastProber struct {
	c        *client
	attempts int
	interval time.Duration
	msgCh    chan *msgAddr
}

func newMulticastProber(cfg *ProbeConfig) (*multicastProber, error) {
	c, err := newClient(!cfg.DisableIPv4, !cfg.DisableIPv6)
	if err != nil {
		return nil, err
	}
	if cfg.Iface != nil {
		if err := c.setInterface(cfg.Iface); err != nil {
			c.Close()
			return nil, err
		}
	}

	msgCh := make(chan *msgAddr, 32)
	if c.use_ipv4 {
		go c.recv(c.ipv4UnicastConn, msgCh)
		go c.recv(c.ipv4MulticastConn, msgCh)
	}
	if c.use_ipv6 {
		go c.recv(c.ipv6UnicastConn, msgCh)
		go c.recv(c.ipv6MulticastConn, msgCh)
	}

	return &multicastProber{
		c:        c,
		attempts: cfg.Attempts,
		interval: cfg.Interval,
		msgCh:    msgCh,
	}, nil
}

// Close releases the underlying client.
func (p *multicastProber) Close() error {
	return p.c.Close()
}

// probeName sends up to Attempts probe queries for name, waiting Interval
// between them, and returns true as soon as a conflicting response is seen.
func (p *multicastProber) probeName(ctx context.Context, name string) (bool, error) {
	q := new(dns.Msg)
	// Probe with QTYPE ANY per RFC 6762 section 8.1 so any record type held by
	// another responder counts as a conflict.
	q.SetQuestion(name, dns.TypeANY)
	q.RecursionDesired = false

	for i := 0; i < p.attempts; i++ {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if err := p.c.sendQuery(q); err != nil {
			return false, err
		}

		deadline := time.NewTimer(p.interval)
		for waiting := true; waiting; {
			select {
			case <-ctx.Done():
				deadline.Stop()
				return false, ctx.Err()
			case resp := <-p.msgCh:
				if msgConflictsWith(resp.msg, name) {
					deadline.Stop()
					return true, nil
				}
				// Unrelated traffic; keep waiting out this interval.
			case <-deadline.C:
				waiting = false
			}
		}
	}
	return false, nil
}
