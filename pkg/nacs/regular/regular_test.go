package regular_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/nacs/regular"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestRegular(t *testing.T) {
	filePath := "../findercache"
	configor, err := regular.NewFileConfigor(filePath, regular.WithSuffix(".findercache"))
	z.Must(err)
	defer configor.Close()

	if err := configor.SetConfig("key1.findercache", []byte("value1")); err != nil {
		panic(err)
	}

	value, err := configor.GetConfig("key1.findercache")
	z.Must(err)
	assert.Equal(t, "value1", string(value))

	ctx, cancel := context.WithCancel(context.Background())
	err = configor.WatchConfig(func(key string, newValue []byte, err error) {
		assert.Equal(t, key, "key1.findercache")
		assert.Equal(t, "value2", string(newValue))
		assert.NoError(t, err)
		cancel()
	})
	z.Must(err)

	time.Sleep(time.Millisecond * 500)
	os.WriteFile("../findercache/key1.findercache", []byte("value2"), os.ModePerm)
	<-ctx.Done()
}
