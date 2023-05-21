package netx_test

import (
	"fmt"
	"testing"

	"github.com/cocktail828/go-tools/netx"
	"github.com/sirupsen/logrus"
)

func TestXxx(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	rrs := &netx.RRSet{Logger: logger, CacheFile: "dns.cache"}

	f := func(addr string) {
		rrs.Refresh(addr)
		fmt.Printf("%#v\n\n", rrs.Endpoints())
	}

	f("baidu.com:8080,qq.com:9090,10.1.98.0:888")
	f("dns://qq.com")
	f("dns+srv://aliyun.com")
}
