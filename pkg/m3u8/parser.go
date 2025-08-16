package m3u8

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Segment represents a single video segment
type Segment struct {
	URL      string
	Filename string
}

// TODO: !!! ASAP !!! Add support for skipping segments
// parseM3U8 parses m3u8 content and returns a list of segments
func parseM3U8(data, forceURLPrefix, forceExt string, skip, limit int) ([]Segment, error) {
	if data == "" {
		return nil, fmt.Errorf("m3u8 data is empty")
	}

	segments := make([]Segment, 0, limit)

	lines := strings.Split(data, "\n")
	segmentNum := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if segmentNum < skip {
			segmentNum++
			continue
		}

		if segmentNum >= skip+limit {
			break
		}

		url, err := url.Parse(forceURLPrefix)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL prefix: %w", err)
		}
		url = url.JoinPath(line)

		ext := filepath.Ext(line)
		if forceExt != "" {
			ext = forceExt
		}

		filename := fmt.Sprintf("segment_%04d%s", segmentNum, ext)
		segments = append(segments, Segment{
			URL:      url.String(),
			Filename: filename,
		})
		segmentNum++
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found")
	}

	return segments, nil
}

// LoadHeadersFromFile loads custom HTTP headers from a JSON file
func LoadHeadersFromFile(headersFile string) (map[string]string, error) {
	if headersFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(headersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers file: %w", err)
	}

	var headers map[string]string

	if len(data) == 0 {
		return headers, nil
	}

	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers file: %w", err)
	}

	return headers, nil
}
