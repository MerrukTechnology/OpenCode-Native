package logging

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MerrukTechnology/OpenCode-Native/internal/pubsub"
	"github.com/go-logfmt/logfmt"
)

const (
	persistKeyArg  = "$_persist"
	PersistTimeArg = "$_persist_time"
)

// LogData holds log messages and provides pub/sub functionality for log events.
type LogData struct {
	messages []LogMessage
	*pubsub.Broker[LogMessage]
	lock sync.Mutex
}

// Add adds a log message to the LogData and publishes it to subscribers.
func (l *LogData) Add(msg LogMessage) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.messages = append(l.messages, msg)
	l.Publish(pubsub.CreatedEvent, msg)
}

// List returns all log messages in the LogData.
func (l *LogData) List() []LogMessage {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.messages
}

var defaultLogData = &LogData{
	messages: make([]LogMessage, 0),
	Broker:   pubsub.NewBroker[LogMessage](),
}

type writer struct{}

func (w *writer) Write(p []byte) (int, error) {
	d := logfmt.NewDecoder(bytes.NewReader(p))

	for d.ScanRecord() {
		msg := LogMessage{
			ID:   strconv.FormatInt(time.Now().UnixNano(), 10),
			Time: time.Now(),
		}
		for d.ScanKeyval() {
			switch string(d.Key()) {
			case "time":
				parsed, err := time.Parse(time.RFC3339, string(d.Value()))
				if err != nil {
					return 0, fmt.Errorf("parsing time: %w", err)
				}
				msg.Time = parsed
			case "level":
				msg.Level = strings.ToLower(string(d.Value()))
			case "msg":
				msg.Message = string(d.Value())
			default:
				if string(d.Key()) == persistKeyArg {
					msg.Persist = true
				} else if string(d.Key()) == PersistTimeArg {
					parsed, err := time.ParseDuration(string(d.Value()))
					if err != nil {
						continue
					}
					msg.PersistTime = parsed
				} else {
					msg.Attributes = append(msg.Attributes, Attr{
						Key:   string(d.Key()),
						Value: string(d.Value()),
					})
				}
			}
		}
		defaultLogData.Add(msg)
	}
	if d.Err() != nil {
		return 0, d.Err()
	}
	return len(p), nil
}

// NewWriter creates a new writer instance for parsing logfmt format.
func NewWriter() *writer {
	w := &writer{}
	return w
}

// Subscribe returns a channel of log messages for real-time log monitoring.
func Subscribe(ctx context.Context) <-chan pubsub.Event[LogMessage] {
	return defaultLogData.Subscribe(ctx)
}

// List returns all log messages from the default LogData.
func List() []LogMessage {
	return defaultLogData.List()
}
