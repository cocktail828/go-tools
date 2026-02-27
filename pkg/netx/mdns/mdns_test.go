package mdns

import (
	"context"
	"testing"
	"time"
)

func TestMDNS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go Announce(ctx, Instance{
		Name:    "_fnos",
		Service: "_mediacenter._udp",
		Port:    8000,
		Info:    []string{"xxx", "yyy"},
	})

	time.Sleep(time.Millisecond * 300)
	entries, err := Lookup(LookParam{
		Service: "_mediacenter._udp",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		t.Logf("%#v\n", entry)
	}
	cancel()
}
