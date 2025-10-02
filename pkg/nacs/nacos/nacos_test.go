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
	uri = "nacos://nacos:nacos@172.29.231.108:8848/group/asfd/v1.0.0?namespace=xxx"
)

func init() {
	_u, err := url.ParseRequestURI(uri)
	z.Must(err)

	_nac, err := NewNacosClient(_u)
	z.Must(err)
	nac = _nac
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
	_, err := nac.Register(nacs.RegisterInstance{
		Address:  "127.0.0.1:8080",
		Metadata: map[string]string{"a": "b"},
	})
	z.Must(err)

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := nac.Watch(nacs.Service{
		Group: nac.Group,
		Name:  nac.ServiceName(),
	}, func(insts []nacs.Instance, err error) {
		t.Logf("watch callback err=%v insts=%v", err, insts)
		f()
	})
	z.Must(err)
	defer cancel()

	time.Sleep(time.Second * 2)
	insts, err := nac.Discover(nacs.Service{
		Group: nac.Group,
		Name:  nac.ServiceName(),
	})
	z.Must(err)
	t.Logf("discover insts=%v", insts)

	<-ctx.Done()
}
