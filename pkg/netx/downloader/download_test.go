package downloader_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/pkg/netx/downloader"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	dl := downloader.Downloader{
		Client:         http.DefaultClient,
		MaxConcurrency: 10,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	reader, err := dl.Parallel(ctx, "https://ddmedia-test.oss-cn-beijing.aliyuncs.com/ddmedia/test/mts/2024_06_05/3602013791191589/101/bfc939d1-2310-11ef-b243-06e43602a2c3.mp3")
	z.Must(err)
	data, err := io.ReadAll(reader)
	z.Must(err)
	assert.Equal(t, 4664802, len(data))
}
