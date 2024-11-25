package m3u8

import "fmt"

// Config holds global configuration
type Config struct {
	verbose bool
}

var config Config

// Vprint prints a message if verbose mode is enabled
func Vprint(format string, a ...interface{}) {
	if config.verbose {
		fmt.Printf(format+"\n", a...)
	}
}

// SetVerbose sets the verbose mode for logging
func SetVerbose(verbose bool) {
	config.verbose = verbose
}
