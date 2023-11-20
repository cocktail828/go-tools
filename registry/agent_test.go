package registry_test

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

func TestProxy_consul(t *testing.T) {
	go http.ListenAndServe(":18081", nil)
	time.Sleep(time.Second)

	c := consulagent.New("127.0.0.1:8500")
	reg := registry.New(
		registry.WithConfiger(c),
		registry.WithRegister(c),
	)
	z.Must(reg.Register(context.Background(), registry.Registration{
		Name:    "demo",
		Version: "1.0.0",
		Address: "127.0.0.1",
		Port:    18081,
	}))
	defer func() {
		z.Must(reg.DeRegister(context.Background()))
	}()

	svcs, err := reg.Services(context.Background(), "demo", "1.0.0")
	z.Must(err)
	fmt.Println("svcs:", svcs)

	fmt.Println(reg.GetConfig(context.Background(), "demo", "1.0.0"))
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)
	<-ctx.Done()
}

func TestCall(t *testing.T) {
	// go http.ListenAndServe(":18081", nil)
	// time.Sleep(time.Second)

	c := consulagent.New("127.0.0.1:8500")
	reg := registry.New(
		registry.WithConfiger(c),
		registry.WithRegister(c),
		registry.WithCaller(func(ctx context.Context, e registry.Entry, b []byte) ([]byte, error) {
			fmt.Println(e)
			return nil, nil
		}),
	)
	// z.Must(reg.Register(context.Background(), registry.Registration{
	// 	Name:    "demo",
	// 	Version: "1.0.0",
	// 	Address: "127.0.0.1",
	// 	Port:    18081,
	// }))
	// defer func() {
	// 	z.Must(reg.DeRegister(context.Background()))
	// }()
	time.Sleep(time.Second)
	_, err := reg.Call(context.Background(), "demo", "1.0.0", nil)
	assert.Equal(t, nil, err)
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
}
