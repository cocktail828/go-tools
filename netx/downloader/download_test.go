package downloader_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cocktail828/go-tools/netx/downloader"
	"github.com/cocktail828/go-tools/z"
	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	dl := downloader.FileDownloader{
		Client:         http.DefaultClient,
		MaxConcurrency: 10,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	buffer, err := dl.Download(ctx, "https://ddmedia-test.oss-cn-beijing.aliyuncs.com/ddmedia/test/mts/2024_06_05/3602013791191589/101/bfc939d1-2310-11ef-b243-06e43602a2c3.mp3")
	z.Must(err)
	assert.Equal(t, 4664802, buffer.Len())
}
