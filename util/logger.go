package util

import (
	"fmt"
	"time"
)

// Logger is a strings-tagged, severity-aware logger interface.
type Logger interface {
	// Debug logs a message with a "debug" severity.
	Debug(format string, a ...interface{})
	// Info logs a message with an "info" severity.
	Info(format string, a ...interface{})
	// Error logs a message with an "error" severity.
	Error(format string, a ...interface{})
	// WithTag returns a copy of the Logger with "tag" added to the existing
	// tags.
	WithTag(tag string) Logger
	// Sync writes pending messages out (flushes message buffers, if any).
	Sync()
}

// Tags implements string tags. To be embedded in Logger interface
// implementations. Implements Stringer.
type Tags struct {
	tags string
}

// With returns a copy of the Tags struct with "tag" added to the existing tags.
func (t Tags) With(tag string) Tags {
	if t.tags == "" {
		return Tags{tag}
	}
	return Tags{t.tags + " " + tag}
}

func (t Tags) String() string {
	return t.tags
}

// NoOpLogger is a Logger implementation which does not log anything.
type NoOpLogger struct{}

func (l NoOpLogger) Debug(format string, a ...interface{}) {}
func (l NoOpLogger) Info(format string, a ...interface{})  {}
func (l NoOpLogger) Error(format string, a ...interface{}) {}
func (l NoOpLogger) WithTag(tag string) Logger {
	return l
}
func (l NoOpLogger) Sync() {}

// ProductionLogger is a Logger implementation which writes severity, tags and
// message to the console.  Debug severity messages are ignored.
type ProductionLogger struct {
	tags Tags
}

func NewProductionLogger(tags string) Logger {
	return &ProductionLogger{
		tags: Tags{tags},
	}
}

func (l ProductionLogger) header(level string) string {
	return fmt.Sprintf("[%s][%s] ", level, l.tags)
}

func (l ProductionLogger) Debug(format string, a ...interface{}) {}
func (l ProductionLogger) Info(format string, a ...interface{}) {
	fmt.Printf(l.header("INFO ")+format+"\n", a...)
}
func (l ProductionLogger) Error(format string, a ...interface{}) {
	fmt.Printf(l.header("ERROR")+format+"\n", a...)
}
func (l ProductionLogger) WithTag(tag string) Logger {
	return &ProductionLogger{l.tags.With(tag)}
}
func (l ProductionLogger) Sync() {}

// DebugLogger is a Logger implementation which writes time, severity, tags and
// message to the console. Time is written in microseconds modulo 10^8.
type DebugLogger struct {
	tags Tags
}

func NewDebugLogger(tags string) Logger {
	return &DebugLogger{
		tags: Tags{tags},
	}
}

func (l DebugLogger) header(level string) string {
	t := (time.Now().UnixNano() / 1000) % 100000000
	return fmt.Sprintf("%08d [%s][%s] ", t, level, l.tags)
}

func (l DebugLogger) Debug(format string, a ...interface{}) {
	fmt.Printf(l.header("DEBUG")+format+"\n", a...)
}
func (l DebugLogger) Info(format string, a ...interface{}) {
	fmt.Printf(l.header("INFO ")+format+"\n", a...)
}
func (l DebugLogger) Error(format string, a ...interface{}) {
	fmt.Printf(l.header("ERROR")+format+"\n", a...)
}
func (l DebugLogger) WithTag(tag string) Logger {
	return &DebugLogger{l.tags.With(tag)}
}
func (l DebugLogger) Sync() {}
