package m3u8

import (
	"fmt"
	"os"
	"os/exec"
)

// CombineSegments combines downloaded segments using ffmpeg
func CombineSegments(fileList, outputFile string, ffmpegPath string, removeFileList bool) error {
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
