package m3u8

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// DownloadM3U8 downloads and parses an M3U8 file
func DownloadM3U8(client *http.Client, url, cacheFile, forceURLPrefix, forceExt string) ([]Segment, error) {
	Vprint("Downloading .m3u8")

	if cacheFile != "" {
		if data, err := os.ReadFile(cacheFile); err == nil {
			Vprint("Using cached .m3u8")
			var segments []Segment
			if err := json.Unmarshal(data, &segments); err == nil {
				return segments, nil
			}
		}
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download m3u8: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read m3u8: %w", err)
	}

	segments := parseM3U8(string(data), forceURLPrefix, forceExt)

	if cacheFile != "" {
		Vprint("Caching .m3u8")
		if data, err := json.Marshal(segments); err == nil {
			os.WriteFile(cacheFile, data, 0644)
		}
	}

	return segments, nil
}

// downloadSegment downloads a single segment
func downloadSegment(client *http.Client, segment Segment, segmentsDir string) (bool, string) {
	Vprint("Downloading segment %s...", segment.Filename)

	outPath := filepath.Join(segmentsDir, segment.Filename)

	resp, err := client.Get(segment.URL)
	if err != nil {
		fmt.Printf("Failed to download %s: %v\n", segment.Filename, err)
		return false, outPath
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download %s: status %d\n", segment.Filename, resp.StatusCode)
		return false, outPath
	}

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Failed to create file %s: %v\n", segment.Filename, err)
		return false, outPath
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to write file %s: %v\n", segment.Filename, err)
		return false, outPath
	}

	return true, outPath
}

type BatchResult struct {
	Success bool
	Path    string
}

// DownloadBatch downloads multiple segments concurrently
func DownloadBatch(client *http.Client, segments []Segment, segmentsDir string, concurrency int) []BatchResult {
	results := make([]BatchResult, len(segments))

	sem := NewSemaphore(concurrency)
	var wg sync.WaitGroup

	for i, segment := range segments {
		wg.Add(1)

		go func(segment Segment, i int) {
			defer wg.Done()

			sem.Acquire()
			defer sem.Release()

			success, path := downloadSegment(client, segment, segmentsDir)
			results[i] = BatchResult{
				Success: success,
				Path:    path,
			}
		}(segment, i)
	}

	wg.Wait()
	return results
}

// HeaderMapTransport implements custom header injection
type HeaderMapTransport struct {
	Headers map[string]string
	Base    http.RoundTripper
}

func (t *HeaderMapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}
	return t.Base.RoundTrip(req)
}
