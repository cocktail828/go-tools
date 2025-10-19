package native

import (
	"context"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestNative(t *testing.T) {
	tests := map[string]string{
		"/tmp/test_config.txt": "file:///tmp/test_config.txt",
		"./test_config.txt":    "file://./test_config.txt",
	}

	for _, uri := range tests {
		u, err := url.ParseRequestURI(uri)
		z.Must(err)
		tempFilePath := path.Join(u.Host, u.Path)
		defer os.Remove(tempFilePath)

		data := []byte("hello world")
		z.Must(os.WriteFile(tempFilePath, data, os.ModePerm))

		configor, err := NewFileConfigor(u)
		z.Must(err)
		defer configor.Close()

		value, err := configor.Load()
		z.Must(err)
		assert.Equal(t, string(data), string(value))

		// monitor
		data = []byte("updated_value")
		ctx, cancel := context.WithCancel(context.Background())
		_, err = configor.Monitor(func(name string, payload []byte, err error) {
			assert.NoError(t, err, "Monitor callback should not return error")

			assert.Equal(t, string(data), string(payload))
			cancel()
		})
		z.Must(err)

		time.Sleep(time.Millisecond * 500)
		z.Must(os.WriteFile(tempFilePath, data, os.ModePerm))

		select {
		case <-ctx.Done():
		case <-time.After(time.Second * 2):
			t.Fatal("Monitor did not trigger within timeout period")
		}
	}
}
