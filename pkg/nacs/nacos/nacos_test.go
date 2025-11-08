package nacos

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
	nac *nacosClient
	uri = "nacos://nacos:nacos@127.0.0.1:8848/myns/mygroup/asfd/v1.0.0"
)

func init() {
}

func TestMain(m *testing.M) {
	_u, err := url.ParseRequestURI(uri)
	z.Must(err)

	_nac, err := NewNacosClient(_u)
	z.Must(err)
	nac = _nac
	// m.Run()
}

func TestConfigor(t *testing.T) {
	bs, err := nac.Load()
	z.Must(err)
	assert.NotEqual(t, 0, len(bs))

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := nac.Monitor(func(name string, payload []byte, err error) {
		t.Logf("monitor callback name=%v payload=%v err=%v", name, len(payload), err)
		f()
	})
	z.Must(err)
	defer cancel()
	<-ctx.Done()
}

func TestNaming(t *testing.T) {
	_, err := nac.Register(nacs.Instance{
		Host:     "127.0.0.1",
		Port:     8080,
		Metadata: map[string]string{"a": "b"},
	})
	z.Must(err)

	expect := []nacs.Instance{
		{
			Enable:   true,
			Healthy:  true,
			Service:  "asfd@v1.0.0",
			Host:     "127.0.0.1",
			Port:     8080,
			Metadata: map[string]string{"a": "b"},
		},
	}

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := nac.Watch(func(insts []nacs.Instance, err error) {
		z.Must(err)
		assert.Equal(t, expect, insts)
		f()
	})
	z.Must(err)
	defer cancel()

	time.Sleep(time.Second * 2)
	insts, err := nac.Discover()
	z.Must(err)
	assert.Equal(t, expect, insts)

	<-ctx.Done()
}
