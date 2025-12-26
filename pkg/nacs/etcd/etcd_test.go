package etcd

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

var (
	ncs *EtcdClient
	uri = "etcd://172.29.231.108:2379,172.29.231.108:2380,172.29.231.108:2381?namespace=myns&service=asfd&version=v1.0.0"
)

func TestMain(m *testing.M) {
	_u, err := url.ParseRequestURI(uri)
	z.Must(err)

	_nac, err := NewEtcdClient(_u)
	z.Must(err)
	ncs = _nac
	m.Run()
}

func TestConfigor(t *testing.T) {
	_, err := ncs.Load()
	z.Must(err)

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := ncs.Monitor(func(name string, payload []byte, err error) {
		t.Logf("monitor callback name=%v payload=%v err=%v", name, len(payload), err)
		f()
	})
	z.Must(err)
	defer cancel()
	<-ctx.Done()
}

func TestNaming(t *testing.T) {
	_, err := ncs.Register("127.0.0.1", 8080, map[string]string{"a": "b"})
	z.Must(err)

	expect := []nacs.Instance{
		{
			Name: "asfd@v1.0.0",
			Host: "127.0.0.1",
			Port: 8080,
			Meta: map[string]string{"a": "b"},
		},
	}

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := ncs.Watch(func(insts []nacs.Instance, err error) {
		z.Must(err)
		assert.Equal(t, expect, insts)
		f()
	})
	z.Must(err)
	defer cancel()

	time.Sleep(time.Second * 2)
	insts, err := ncs.Discover()
	z.Must(err)
	assert.Equal(t, expect, insts)

	<-ctx.Done()
}
