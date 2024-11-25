package m3u8

import "fmt"

// Logger defines the interface for logging operations
type Logger interface {
	Printf(format string, args ...interface{})
}

// Config holds configuration for m3u8 operations
type Config struct {
	Logger  Logger
	Verbose bool
}

// DefaultConfig returns a new Config with default settings
func DefaultConfig() *Config {
	return &Config{
		Logger:  defaultLogger{},
		Verbose: false,
	}
}

type defaultLogger struct{}

func (l defaultLogger) Printf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
