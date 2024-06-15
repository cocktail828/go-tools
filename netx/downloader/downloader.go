package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	DefaultMinSize = 5 * 1024 * 1024 // object bigger than this will be downloaded parallely
)

// FileDownloader 文件下载器
type FileDownloader struct {
	Client         *http.Client
	MaxConcurrency int // 最大并发
	MinSize        int // 触发并发的最小大小
}

func (dl *FileDownloader) Download(ctx context.Context, url string) (*bytes.Buffer, error) {
	minsize := dl.MinSize
	if minsize <= 0 {
		minsize = DefaultMinSize
	}

	if dl.MaxConcurrency <= 1 {
		return dl.Single(ctx, url)
	}

	size, err := dl.AllowRange(ctx, url)
	if err != nil || minsize > size {
		return dl.Single(ctx, url)
	}

	return dl.Multi(ctx, size, url)
}

func (dl *FileDownloader) Single(ctx context.Context, url string) (*bytes.Buffer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dl.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(body), nil
}

// head 获取要下载的文件的基本信息(header) 使用HTTP Method Head
func (dl *FileDownloader) AllowRange(ctx context.Context, url string) (int, error) {
	r, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := dl.Client.Do(r)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("range requests disallow for: http.Status=%v", resp.StatusCode)
	}

	//检查是否支持 断点续传
	//https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Ranges
	if strings.ToLower(resp.Header.Get("Accept-Ranges")) != "bytes" {
		return 0, errors.Errorf("range requests disallow for: Accept-Ranges != bytes")
	}

	//https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Length
	return strconv.Atoi(resp.Header.Get("Content-Length"))
}

// Download 开始下载任务
func (dl *FileDownloader) Multi(ctx context.Context, size int, url string) (*bytes.Buffer, error) {
	partNum := dl.MaxConcurrency
	if partNum <= 0 {
		partNum = 1
	}
	partSize := size / partNum

	eg := errgroup.Group{}
	eg.SetLimit(dl.MaxConcurrency)
	parts := []*DownloadPart{}
	for i := 0; i < partNum; i++ {
		t := &DownloadPart{
			Client: dl.Client,
			Url:    url,
			Start:  partSize * i,
			End: func() int {
				if (i + 1) == partNum {
					return size - 1
				}
				return partSize*(i+1) - 1
			}(),
		}
		parts = append(parts, t)
		eg.Go(func() error { return t.Download(ctx) })
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(make([]byte, 0, size))
	for _, part := range parts {
		buffer.Write(part.data)
	}
	return buffer, nil
}

// filePart 文件分片
type DownloadPart struct {
	Client     *http.Client
	Url        string
	Start, End int
	data       []byte
}

// 下载分片
func (dl *DownloadPart) Download(ctx context.Context) error {
	r, err := http.NewRequestWithContext(ctx, "GET", dl.Url, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", dl.Start, dl.End))

	resp, err := dl.Client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return errors.Errorf("download fail for: http.Status=%v", resp.StatusCode)
	}

	dl.data, err = io.ReadAll(resp.Body)
	return err
}

func (dl *DownloadPart) Data() []byte { return dl.data }
