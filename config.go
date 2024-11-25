package main

import "fmt"

// Config holds global configuration
type Config struct {
	verbose bool
}

var config Config

// vprint prints a message if verbose mode is enabled
func vprint(format string, a ...interface{}) {
	if config.verbose {
		fmt.Printf(format+"\n", a...)
	}
}
