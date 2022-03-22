package util

import (
	"fmt"
	"log/syslog"
)

// SyslogLogger writes log messages to the syslog using /dev/log socket.
// Debug severity messages are written optionally.
type SyslogLogger struct {
	syslog *syslog.Writer
	tags   Tags
	debug  bool
}

// NewSyslogLogger creates a new SyslogLogger. "debug" parameter determines whether
// debug severity messages should be logged or not.
func NewSyslogLogger(tag string, debug bool) (Logger, error) {
	const syslogTag = "mqtt_sn_gateway"
	const device = "/dev/log"

	syslog, err := syslog.Dial("unixgram", device, syslog.LOG_DAEMON, syslogTag)
	if err != nil {
		return nil, err
	}

	logger := &SyslogLogger{syslog: syslog, tags: Tags{tag}, debug: debug}

	return logger, nil
}

func (l *SyslogLogger) Debug(format string, a ...interface{}) {
	if l.debug {
		l.syslog.Debug(fmt.Sprintf(format, a...))
	}
}

func (l *SyslogLogger) Info(format string, a ...interface{}) {
	l.syslog.Info(fmt.Sprintf(format, a...))
}

func (l *SyslogLogger) Error(format string, a ...interface{}) {
	l.syslog.Err(fmt.Sprintf(format, a...))
}

func (l *SyslogLogger) WithTag(tag string) Logger {
	return &SyslogLogger{
		syslog: l.syslog,
		tags:   l.tags.With(tag),
		debug:  l.debug,
	}
}

func (l *SyslogLogger) Sync() {}

func (l *SyslogLogger) header(level string) string {
	return fmt.Sprintf("[%s] ", l.tags)
}
