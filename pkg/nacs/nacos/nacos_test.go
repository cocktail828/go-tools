package nacos

import (
	"context"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/z"
)

var (
	nac     *nacosClient
	cfgname = "asfd"
	svcName = "asfd@v1.0.0"
)

func init() {
	_nac, err := NewNacosClient("nacos://nacos:nacos@172.29.231.108:8848?appname=xxx")
	z.Must(err)
	nac = _nac
}

func TestConfigor(t *testing.T) {
	bs, err := nac.Load(WithLoadID(cfgname))
	z.Must(err)
	t.Logf("load config err=%v len(cfg)=%d", err, len(bs))

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := nac.Monitor(func(err error, args ...any) {
		t.Logf("monitor callback err=%v args=%v", err, args)
		f()
	}, WithMonitorID(cfgname))
	z.Must(err)
	defer cancel()
	cancel()
	<-ctx.Done()
}

func TestNaming(t *testing.T) {
	_, err := nac.Register(nacs.RegisterInstance{
		Name:     svcName,
		Address:  "127.0.0.1:8080",
		Metadata: map[string]string{"a": "b"},
	})
	z.Must(err)

	ctx, f := context.WithTimeout(context.Background(), time.Second*3)
	cancel, err := nac.Watch(nacs.Service{
		Name: svcName,
	}, func(insts []nacs.Instance, err error) {
		t.Logf("watch callback err=%v insts=%v", err, insts)
		f()
	})
	z.Must(err)
	defer cancel()

	time.Sleep(time.Second * 2)
	_, err = nac.Discover(nacs.Service{Name: svcName})
	z.Must(err)

	<-ctx.Done()
}
