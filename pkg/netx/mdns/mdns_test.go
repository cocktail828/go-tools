package mdns

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMDNS(t *testing.T) {
	svc := &Service{
		Name:    "_fnos",
		Service: "_nas._tcp",
		Port:    80,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go svc.Register(ctx)

	time.Sleep(time.Millisecond * 300)
	entries, err := Lookup(context.Background(), "_nas._tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Println(entry)
	}
	cancel()
}
