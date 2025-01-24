package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cocktail828/go-tools/pkg/retry"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	DefaultMinSize = 5 * 1024 * 1024 // Default minimum file size to trigger parallel download
)

// Downloader is a file downloader that supports both single and parallel downloads.
type Downloader struct {
	Client         *http.Client
	MaxConcurrency int // Maximum number of concurrent downloads
	SizeThreshold  int // Minimum file size to trigger parallel download
}

func (dl *Downloader) init() {
	if dl.Client == nil {
		dl.Client = http.DefaultClient
	}
}

// GetFileSize retrieves the file size and checks if the server supports range requests.
func (dl *Downloader) GetFileSize(ctx context.Context, url string) (int, error) {
	dl.init()
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create HEAD request")
	}

	resp, err := dl.Client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute HEAD request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("request failed: HTTP status %d", resp.StatusCode)
	}

	// Check if the server supports range requests
	if strings.ToLower(resp.Header.Get("Accept-Ranges")) != "bytes" {
		return 0, errors.New("server does not support range requests")
	}

	// Get the file size
	contentLength := resp.Header.Get("Content-Length")
	size, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse Content-Length")
	}

	return size, nil
}

// download performs a generic download operation with a given range.
func (dl *Downloader) download(ctx context.Context, url string, start, end int) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GET request")
	}

	if start >= 0 && end >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	}

	resp, err := dl.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute GET request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, errors.Errorf("download failed: HTTP status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return bytes.NewReader(data), nil
}

// Sequential downloads a file sequentially (single-threaded).
func (dl *Downloader) Sequential(ctx context.Context, url string) (io.Reader, error) {
	dl.init()
	return dl.download(ctx, url, -1, -1)
}

// Parallel downloads a file in parallel using multiple goroutines.
func (dl *Downloader) Parallel(ctx context.Context, url string) (io.Reader, error) {
	dl.init()

	// Set the minimum file size threshold
	minSize := dl.SizeThreshold
	if minSize <= 0 {
		minSize = DefaultMinSize
	}

	// Fallback to sequential download if MaxConcurrency is 1 or less
	if dl.MaxConcurrency <= 1 {
		return dl.Sequential(ctx, url)
	}

	// Get the file size and check if parallel download is supported
	size, err := dl.GetFileSize(ctx, url)
	if err != nil || minSize > size {
		return dl.Sequential(ctx, url)
	}

	// Calculate the size of each part
	partNum := dl.MaxConcurrency
	partSize := size / partNum

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(dl.MaxConcurrency)

	// Download parts in parallel
	parts := make([]io.Reader, partNum)
	for i := 0; i < partNum; i++ {
		i := i // Avoid closure capture issue
		start := partSize * i
		end := start + partSize - 1
		if i == partNum-1 {
			end = size - 1 // The last part includes the remaining data
		}

		eg.Go(func() error {
			return retry.Do(func() error {
				reader, err := dl.download(ctx, url, start, end)
				if err != nil {
					return errors.Wrapf(err, "failed to download part %d", i)
				}
				parts[i] = reader
				return nil
			}, retry.Attempts(3))
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, errors.Wrap(err, "parallel download failed")
	}

	return io.MultiReader(parts...), nil
}
