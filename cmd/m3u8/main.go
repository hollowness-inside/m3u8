package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hollowness-inside/m3u8/pkg/m3u8"
	"github.com/spf13/cobra"
)

var (
	segmentsDir    string // directory for downloaded segments
	forceExt       string // override segment file extensions
	forceURLPrefix string // prefix to prepend to segment URLs
	cacheFile      string // path to segment cache file
	fileList       string // file containing segments to combine
	combine        string // output path for combined segments
	forceCombine   string // combine segments even if some are missing
	cleanup        bool   // delete segments after combining
	fix            string // directory with partial downloads to fix
	headersFile    string // custom HTTP headers file
	skip           int    // number of segments to skip
	limit          int    // maximum segments to download
	concurrent     int    // maximum concurrent downloads
	ffmpegPath     string // path to ffmpeg binary
)

// runE is executed as the main entry point for the Cobra command
func runE(cmd *cobra.Command, args []string) error {
	url := args[0]
	ctx := context.Background()

	// Load headers from file if specified
	headers, err := m3u8.LoadHeadersFromFile(headersFile)
	if err != nil {
		return err
	}

	// Create downloader
	downloader := m3u8.NewDownloader()
	downloader.SetHeaders(headers)
	downloader.SetCacheFile(cacheFile)

	// Handle fix mode, where segments are already partially downloaded
	if fix != "" {
		// Check if the fix directory exists
		if _, err := os.Stat(fix); os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist", fix)
		}

		// Determine the file extension from the first file in the directory if not forced
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
		}

		// Set the segments directory to the fix directory
		segmentsDir = fix
	} else {
		// Create the segments directory if it doesn't exist
		if err := os.MkdirAll(segmentsDir, 0755); err != nil {
			return fmt.Errorf("failed to create segments directory: %w", err)
		}
	}

	// Download and parse the M3U8 file to get the list of segments
	segments, err := downloader.DownloadM3U8(ctx, url, forceURLPrefix, forceExt, skip, limit)
	if err != nil {
		return err
	}

	// Filter out already downloaded segments in fix mode
	if fix != "" {
		existingSegments := make(map[string]struct{})
		files, _ := os.ReadDir(segmentsDir)
		for _, file := range files {
			if !file.IsDir() {
				name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				existingSegments[name] = struct{}{}
			}
		}

		var missingSegments []m3u8.Segment
		for _, seg := range segments {
			name := strings.TrimSuffix(seg.Filename, filepath.Ext(seg.Filename))
			if _, exists := existingSegments[name]; !exists {
				missingSegments = append(missingSegments, seg)
			}
		}

		// If no segments are missing, exit early
		if len(missingSegments) == 0 {
			fmt.Println("All segments are already downloaded")
			return nil
		}

		fmt.Printf("Found %d segments to fix\n", len(missingSegments))
		segments = missingSegments
	}

	// Download segments concurrently
	results := downloader.DownloadBatch(ctx, segments, segmentsDir, concurrent)

	// Count successful downloads
	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
	}

	// Report download results
	if successCount == len(segments) {
		fmt.Println("All segments downloaded successfully!")
	} else {
		fmt.Printf("Failed to download %d out of %d segments\n",
			len(segments)-successCount, len(segments))
	}

	// Determine the output file for combination
	outputFile := forceCombine
	if outputFile == "" {
		outputFile = combine
	}

	// If no output file specified or forced combination fails, exit
	if outputFile == "" || (forceCombine != "" && successCount != len(segments)) {
		return nil
	}

	// Create filelist for ffmpeg
	if err := createFileList(fileList, results); err != nil {
		return fmt.Errorf("failed to create filelist: %w", err)
	}

	// Combine segments into a single output file
	if err := m3u8.CombineSegments(fileList, outputFile, ffmpegPath, cleanup); err != nil {
		return fmt.Errorf("failed to combine segments: %w", err)
	}

	// Cleanup segments directory if requested
	if cleanup {
		fmt.Printf("Cleaning up segments directory \"%s\"...\n", segmentsDir)
		os.RemoveAll(segmentsDir)
	}

	return nil
}

// createFileList creates a file containing paths to successfully downloaded segments
func createFileList(fileList string, results []m3u8.BatchResult) error {
	f, err := os.Create(fileList)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, result := range results {
		if result.Error == nil {
			fmt.Fprintf(f, "file '%s'\n", result.Path)
		}
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "m3u8_download",
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
	flags.StringVar(&headersFile, "headers", "", "Path to JSON file containing request headers")
	flags.IntVar(&skip, "skip", 0, "Skip the first N segments")
	flags.IntVar(&limit, "limit", 0, "Limit the number of segments to download")
	flags.IntVar(&concurrent, "concurrent", 10, "Number of concurrent downloads")
	flags.StringVar(&ffmpegPath, "ffmpeg", "", "Path to ffmpeg executable")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
