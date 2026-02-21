package logging

import (
	"time"
)

// LogMessage represents a log message with metadata for the logging system.
type LogMessage struct {
	ID          string
	Time        time.Time
	Level       string
	Persist     bool          // used when we want to show the message in the status bar
	PersistTime time.Duration // used when we want to show the message in the status bar
	Message     string        `json:"msg"`
	Attributes  []Attr
}

// Attr represents a key-value pair for log message attributes.
type Attr struct {
	Key   string
	Value string
}
