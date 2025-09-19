package nacos

import (
	"testing"

	"github.com/cocktail828/go-tools/z"
)

func TestConfigor(t *testing.T) {
	configor, err := NewNacosClient("nacos://172.29.231.108:8848?namespace=public")
	z.Must(err)
	defer configor.Close()

	value, err := configor.Load(WithLoadID("asfd"))
	z.Must(err)
	t.Logf("value: %s", value)
}
