package consulagent_test

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/registry"
	"github.com/cocktail828/go-tools/registry/consulagent"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

var (
	service = "demo"
	version = "1.0.0"
)

func TestConsul(t *testing.T) {
	go http.ListenAndServe(":18081", nil)
	time.Sleep(time.Second)
	c := consulagent.New("127.0.0.1:8500")
	dreg, err := c.Register(context.Background(), registry.Registration{
		Name:    service,
		Version: version,
		Address: "127.0.0.1",
		Port:    18081,
	})
	z.Must(err)
	defer dreg.DeRegister(context.Background())

	entries, err := c.Services(context.Background(), service, version)
	assert.Equal(t, nil, err)
	z.Must(err)
	assert.Equal(t, 1, len(entries))

	go func() {
		assert.Equal(t, nil, c.WatchServices(context.Background(), func(entries []registry.Entry) {
			fmt.Println("svcsss:", entries)
		}))
	}()

	go func() {
		assert.Equal(t, nil, c.WatchService(context.Background(), service, version, func(entries []registry.Entry) {
			fmt.Println("svc:", entries)
		}))
	}()

	// config
	fmt.Println("config...")
	body, err := c.Pull(context.Background(), service, version)
	assert.Equal(t, nil, err)
	for k, v := range body {
		fmt.Println("Pull:", k, string(v))
	}

	go func() {
		c.WatchConfig(context.Background(), service, version, func(kvs map[string][]byte) {
			for k, v := range kvs {
				fmt.Println("kv:", k, string(v))
			}
		})
	}()

	// event
	assert.Equal(t, nil, c.Fire(context.Background(), service, version, "xxx", registry.Event{}))
	l, err := c.Recv(context.Background(), "xxx")
	assert.Equal(t, nil, err)
	fmt.Println("ev:", len(l), l)
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	<-ctx.Done()
}
