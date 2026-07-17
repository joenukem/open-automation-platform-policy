// Package audit records policy decisions as normalized events: an in-memory
// ring buffer for the /v1/policy/decisions API plus JSON-line emission to stdout
// so a platform log/observability collector can ingest them.
package audit

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// Event is one recorded policy decision, in the normalized AutomationEvent shape
// the observability layer consumes.
type Event struct {
	Timestamp     string `json:"timestamp"`
	Service       string `json:"service"`
	EventType     string `json:"event_type"`
	Action        string `json:"action"`
	Actor         string `json:"actor"`
	Org           string `json:"org,omitempty"`
	Outcome       string `json:"outcome"`
	MatchedPolicy string `json:"matched_policy,omitempty"`
	Reason        string `json:"reason,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// Log is a bounded, concurrency-safe decision log.
type Log struct {
	mu   sync.Mutex
	buf  []Event
	max  int
	sink io.Writer
	now  func() time.Time
}

// NewLog returns a Log that keeps the last max events and also writes each event
// as a JSON line to sink (nil disables sink emission).
func NewLog(max int, sink io.Writer) *Log {
	if max <= 0 {
		max = 1000
	}
	return &Log{max: max, sink: sink, now: time.Now}
}

// Record appends an event, trims to the bound, and emits a JSON line to the sink.
func (l *Log) Record(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Timestamp == "" {
		e.Timestamp = l.now().UTC().Format(time.RFC3339Nano)
	}
	if e.Service == "" {
		e.Service = "policy"
	}
	if e.EventType == "" {
		e.EventType = "policy.decision"
	}
	l.buf = append(l.buf, e)
	if len(l.buf) > l.max {
		l.buf = l.buf[len(l.buf)-l.max:]
	}
	if l.sink != nil {
		if b, err := json.Marshal(e); err == nil {
			l.sink.Write(append(b, '\n'))
		}
	}
}

// Recent returns up to n most-recent events, newest last.
func (l *Log) Recent(n int) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || n > len(l.buf) {
		n = len(l.buf)
	}
	out := make([]Event, n)
	copy(out, l.buf[len(l.buf)-n:])
	return out
}
