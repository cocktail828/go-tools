package regular_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/nacs/configuration"
	"github.com/cocktail828/go-tools/pkg/nacs/impl/regular"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestRegular(t *testing.T) {
	rootdir := os.TempDir()
	configor, err := regular.NewFileConfigor(rootdir, regular.WithSuffix(".findercache"))
	z.Must(err)
	defer configor.Close()

	assert.NoError(t, configor.Set(configuration.Config{ID: "key1.findercache"}, []byte("value1")))

	value, err := configor.Get(configuration.Config{ID: "key1.findercache"})
	z.Must(err)
	assert.Equal(t, "value1", string(value))

	ctx, cancel := context.WithCancel(context.Background())
	_, err = configor.Monitor(configuration.Config{ID: "key1.findercache"}, func(cfg configuration.Config, newValue []byte, err error) {
		assert.Equal(t, "key1.findercache", cfg.ID)
		assert.Equal(t, "value2", string(newValue))
		assert.NoError(t, err)
		cancel()
	})
	z.Must(err)

	time.Sleep(time.Millisecond * 500)
	os.WriteFile(rootdir+"/key1.findercache", []byte("value2"), os.ModePerm)
	<-ctx.Done()
}
