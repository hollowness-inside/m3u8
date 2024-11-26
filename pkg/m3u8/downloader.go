package m3u8

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Downloader handles M3U8 downloads and segment management
type Downloader struct {
	client    *http.Client
	transport *HeaderMapTransport
}

// NewDownloader creates a new downloader with the given configuration
func NewDownloader() *Downloader {
	transport := &HeaderMapTransport{
		Base: http.DefaultTransport,
	}

	return &Downloader{
		client: &http.Client{
			Transport: transport,
		},
		transport: transport,
	}
}

// SetHeaders sets the HTTP headers for requests
func (d *Downloader) SetHeaders(headers map[string]string) {
	if headers == nil {
		return
	}

	d.transport.Headers = headers
}

// DownloadM3U8 downloads and parses an M3U8 file
func (d *Downloader) DownloadM3U8(ctx context.Context, url, cacheFile, forceURLPrefix, forceExt string, skip, limit int) ([]Segment, error) {
	fmt.Printf("Downloading .m3u8")

	if cacheFile != "" {
		segments, err := d.loadCache(cacheFile)
		if err == nil {
			return segments, nil
		}
	}

	segments, err := d.fetchM3U8(ctx, url, forceURLPrefix, forceExt, skip, limit)
	if err != nil {
		return nil, err
	}

	if cacheFile != "" {
		if err := d.saveCache(cacheFile, segments); err != nil {
			return nil, err
		}
	}

	return segments, nil
}

func (d *Downloader) loadCache(cacheFile string) ([]Segment, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Using cached .m3u8")
	var segments []Segment
	if err := json.Unmarshal(data, &segments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached m3u8: %w", err)
	}

	if segments == nil {
		return nil, fmt.Errorf("cached m3u8 is empty")
	}

	return segments, nil
}

func (d *Downloader) saveCache(cacheFile string, segments []Segment) error {
	fmt.Printf("Caching .m3u8")

	data, err := json.Marshal(segments)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func (d *Downloader) fetchM3U8(ctx context.Context, url, forceURLPrefix, forceExt string, skip int, limit int) ([]Segment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download m3u8: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read m3u8: %w", err)
	}

	segments := parseM3U8(string(data), forceURLPrefix, forceExt, skip, limit)
	if segments == nil {
		return nil, fmt.Errorf("failed to parse m3u8 data")
	}

	return segments, nil
}

// BatchResult represents the result of a segment download
type BatchResult struct {
	Index int
	Path  string
	Error error
}

// DownloadBatch downloads multiple segments concurrently
func (d *Downloader) DownloadBatch(ctx context.Context, segments []Segment, segmentsDir string, concurrency int) []BatchResult {
	resultsChan := make(chan BatchResult, len(segments))
	sem := newSemaphore(concurrency)
	var wg sync.WaitGroup

	for i, segment := range segments {
		wg.Add(1)
		go func(segment Segment, i int) {
			defer wg.Done()
			sem.acquire()
			defer sem.release()

			path := filepath.Join(segmentsDir, segment.Filename)
			err := d.downloadSegment(ctx, segment.URL, path)
			resultsChan <- BatchResult{
				Index: i,
				Path:  path,
				Error: err,
			}
		}(segment, i)
	}

	wg.Wait()
	close(resultsChan)

	results := make([]BatchResult, len(segments))
	for result := range resultsChan {
		results[result.Index] = result
	}

	return results
}

func (d *Downloader) downloadSegment(ctx context.Context, url string, path string) error {
	fmt.Printf("Downloading segment %s...\n", filepath.Base(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
