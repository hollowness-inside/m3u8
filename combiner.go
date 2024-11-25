package main

import (
	"fmt"
	"os"
	"os/exec"
)

// combineSegments combines downloaded segments using ffmpeg
func combineSegments(fileList, outputFile string, ffmpegPath string, removeFileList bool) error {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileList,
		"-c", "copy",
		outputFile,
	}

	cmd := exec.Command(ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	if removeFileList {
		os.Remove(fileList)
	}

	return nil
}
