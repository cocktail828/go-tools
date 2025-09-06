package regular

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestRegular(t *testing.T) {
	tempFilePath := "/tmp/test_config.txt"
	defer os.Remove(tempFilePath)

	data := []byte("hello world")
	z.Must(os.WriteFile(tempFilePath, data, os.ModePerm))
	configor, err := NewFileConfigor(tempFilePath)
	z.Must(err)
	defer configor.Close()

	value, err := configor.Get(FileName(tempFilePath))
	z.Must(err)
	assert.Equal(t, string(data), string(value), "Read content should match initial content")

	// monitor
	data = []byte("updated_value")
	ctx, cancel := context.WithCancel(context.Background())
	_, err = configor.Monitor(func(err error) {
		assert.NoError(t, err, "Monitor callback should not return error")

		value, err := configor.Get(FileName(tempFilePath))
		z.Must(err)
		assert.Equal(t, string(data), string(value), "Read content should match initial content")
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
