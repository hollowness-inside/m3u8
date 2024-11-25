package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hollowness-inside/m3u8/pkg/m3u8"
	"github.com/spf13/cobra"
)

var (
	segmentsDir    string
	forceExt       string
	forceURLPrefix string
	cacheFile      string
	fileList       string
	combine        string
	forceCombine   string
	cleanup        bool
	fix            string
	headers        string
	limit          int
	concurrent     int
	ffmpegPath     string
)

func runE(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Configure verbose printing
	m3u8.SetVerbose(cmd.Flags().Changed("verbose"))

	// Load headers
	headerMap, err := m3u8.LoadHeaders(headers)
	if err != nil {
		return err
	}

	// Create HTTP client
	client := &http.Client{}
	if headerMap != nil {
		client.Transport = &m3u8.HeaderMapTransport{
			Headers: headerMap,
			Base:    http.DefaultTransport,
		}
	}

	// Fix mode
	if fix != "" {
		if _, err := os.Stat(fix); os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist", fix)
		}

		// Get extension from first file if not forced
		if forceExt == "" {
			files, err := os.ReadDir(fix)
			if err != nil {
				return fmt.Errorf("failed to read directory: %w", err)
			}

			for _, file := range files {
				if !file.IsDir() {
					forceExt = filepath.Ext(file.Name())
					break
				}
			}

			if forceExt == "" {
				return fmt.Errorf("could not determine segment extension from files")
			}
		}

		segmentsDir = fix
	} else {
		if err := os.MkdirAll(segmentsDir, 0755); err != nil {
			return fmt.Errorf("failed to create segments directory: %w", err)
		}
	}

	// Download and parse M3U8
	segments, err := m3u8.DownloadM3U8(client, url, cacheFile, forceURLPrefix, forceExt)
	if err != nil {
		return err
	}

	// Apply segment limit if specified
	if limit > 0 {
		m3u8.Vprint("Limiting download to first %d segments", limit)
		if limit < len(segments) {
			segments = segments[:limit]
		}
	}

	// Filter out already downloaded segments in fix mode
	if fix != "" {
		existingSegments := make(map[string]bool)
		files, _ := os.ReadDir(segmentsDir)
		for _, file := range files {
			if !file.IsDir() {
				name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				existingSegments[name] = true
			}
		}

		var newSegments []m3u8.Segment
		for _, seg := range segments {
			name := strings.TrimSuffix(seg.Filename, filepath.Ext(seg.Filename))
			if !existingSegments[name] {
				newSegments = append(newSegments, seg)
			}
		}

		if len(newSegments) == 0 {
			fmt.Println("All segments are already downloaded")
			return nil
		}
		m3u8.Vprint("Found %d segments to fix", len(newSegments))
		segments = newSegments
	}

	// Download segments
	results := m3u8.DownloadBatch(client, segments, segmentsDir, concurrent)

	// Count successful downloads
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount == len(segments) {
		fmt.Println("All segments downloaded successfully!")
	} else {
		fmt.Printf("Failed to download %d out of %d segments\n",
			len(segments)-successCount, len(segments))
	}

	// Handle combination
	outputFile := forceCombine
	if outputFile == "" {
		outputFile = combine
	}

	if outputFile == "" || (!cmd.Flags().Changed("force-combine") && successCount != len(segments)) {
		return nil
	}

	// Sort results by segment number
	sort.Slice(results, func(i, j int) bool {
		numI := regexp.MustCompile(`(\d+)`).FindString(filepath.Base(results[i].Path))
		numJ := regexp.MustCompile(`(\d+)`).FindString(filepath.Base(results[j].Path))
		return numI < numJ
	})

	// Create filelist for ffmpeg
	f, err := os.Create(fileList)
	if err != nil {
		return fmt.Errorf("failed to create filelist: %w", err)
	}

	for _, result := range results {
		if result.Success {
			fmt.Fprintf(f, "file '%s'\n", result.Path)
		}
	}
	f.Close()

	// Combine segments
	if err := m3u8.CombineSegments(fileList, outputFile, ffmpegPath, cleanup); err != nil {
		return fmt.Errorf("failed to combine segments: %w", err)
	}

	// Cleanup if requested
	if cleanup {
		m3u8.Vprint("Cleaning up segments directory %s...", segmentsDir)
		os.RemoveAll(segmentsDir)
	}

	return nil
}

func main() {

	rootCmd := &cobra.Command{
		Use:   "m3u8_download [url]",
		Short: "Download and combine M3U8 segments",
		Args:  cobra.ExactArgs(1),
		RunE:  runE,
	}

	flags := rootCmd.Flags()
	flags.StringVar(&segmentsDir, "segments-dir", "segments", "Directory to store segments")
	flags.StringVar(&forceExt, "force-ext", "", "Force specific extension for segments (e.g., .ts)")
	flags.StringVar(&forceURLPrefix, "force-url-prefix", "", "Force URL prefix for segments")
	flags.StringVar(&cacheFile, "cache", "", "Path to cache parsed m3u8")
	flags.StringVar(&fileList, "filelist", "filelist.txt", "Path for ffmpeg filelist")
	flags.StringVar(&combine, "combine", "", "Combine segments into OUTPUT file after download")
	flags.StringVar(&forceCombine, "force-combine", "", "Combine segments even if some failed to download")
	flags.BoolVar(&cleanup, "cleanup", false, "Remove segments directory after successful combination")
	flags.StringVar(&fix, "fix", "", "Fix missing segments in the specified directory")
	flags.BoolP("verbose", "v", false, "Enable verbose output")
	flags.StringVar(&headers, "headers", "", "Path to JSON file containing request headers")
	flags.IntVar(&limit, "limit", 0, "Limit the number of segments to download")
	flags.IntVar(&concurrent, "concurrent", 10, "Number of concurrent downloads")
	flags.StringVar(&ffmpegPath, "ffmpeg", "", "Path to ffmpeg executable")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
