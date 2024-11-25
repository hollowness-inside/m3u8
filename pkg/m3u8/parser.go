package m3u8

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Segment represents a single video segment
type Segment struct {
	URL      string
	Filename string
}

// parseM3U8 parses m3u8 content and returns a list of segments
func parseM3U8(data string, forceURLPrefix, forceExt string) []Segment {
	var segments []Segment
	lines := strings.Split(data, "\n")
	segmentNum := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		url := line
		if forceURLPrefix != "" {
			url = forceURLPrefix + line
		}

		ext := filepath.Ext(line)
		if forceExt != "" {
			ext = forceExt
		}

		filename := fmt.Sprintf("segment_%04d%s", segmentNum, ext)
		segments = append(segments, Segment{
			URL:      url,
			Filename: filename,
		})
		segmentNum++
	}

	return segments
}

// LoadHeaders loads custom HTTP headers from a JSON file
func LoadHeaders(headersFile string) (map[string]string, error) {
	if headersFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(headersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers file: %w", err)
	}

	var headers map[string]string
	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers file: %w", err)
	}

	return headers, nil
}
