package common

import (
	"fmt"
	"io"
	"log"
)

type LogLevel int

const (
	LOG_DEBUG LogLevel = iota
	LOG_INFO
	LOG_WARNING
	LOG_ERR
	LOG_OFF
	LOG_UNKNOWN
)

// default log options
const (
	DEFAULT_LOG_PREFIX = "[microbot]"
	DEFAULT_LOG_FLAG   = log.Ldate | log.Lmicroseconds
	DEFAULT_LOG_LEVEL  = LOG_INFO
)

type DefaultLogger struct {
	DEBUG *log.Logger
	ERR   *log.Logger
	INFO  *log.Logger
	WARN  *log.Logger
	level LogLevel
}

// logger interface
type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Level() LogLevel
	SetLevel(l LogLevel)
}

var _ Logger = &DefaultLogger{}

// NewDefaultLogger use a special io.Writer as logger output
func NewDefaultLogger(out io.Writer) *DefaultLogger {
	return NewDefaultLogger2(out, DEFAULT_LOG_PREFIX, DEFAULT_LOG_FLAG)
}

// NewDefaultLogger2 let you customrize your logger prefix and flag
func NewDefaultLogger2(out io.Writer, prefix string, flag int) *DefaultLogger {
	return NewDefaultLogger3(out, prefix, flag, DEFAULT_LOG_LEVEL)
}

// NewDefaultLogger3 let you customrize your logger prefix and flag and logLevel
func NewDefaultLogger3(out io.Writer, prefix string, flag int, l LogLevel) *DefaultLogger {
	return &DefaultLogger{
		DEBUG: log.New(out, fmt.Sprintf("%s [debug] ", prefix), flag),
		ERR:   log.New(out, fmt.Sprintf("%s [error] ", prefix), flag),
		INFO:  log.New(out, fmt.Sprintf("%s [info]  ", prefix), flag),
		WARN:  log.New(out, fmt.Sprintf("%s [warn]  ", prefix), flag),
		level: l,
	}
}

// Debug implement Logger
func (d *DefaultLogger) Debug(v ...interface{}) {
	if d.level <= LOG_DEBUG {
		d.DEBUG.Output(2, fmt.Sprint(v...))
	}
}

// Debugf implement Logger
func (d *DefaultLogger) Debugf(format string, v ...interface{}) {
	if d.level <= LOG_DEBUG {
		d.DEBUG.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error implement Logger
func (d *DefaultLogger) Error(v ...interface{}) {
	if d.level <= LOG_ERR {
		d.ERR.Output(2, fmt.Sprint(v...))
	}
}

// Errorf implement Logger
func (d *DefaultLogger) Errorf(format string, v ...interface{}) {
	if d.level <= LOG_ERR {
		d.ERR.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info implement Logger
func (d *DefaultLogger) Info(v ...interface{}) {
	if d.level <= LOG_INFO {
		d.INFO.Output(2, fmt.Sprint(v...))
	}
}

// Infof implement Logger
func (d *DefaultLogger) Infof(format string, v ...interface{}) {
	if d.level <= LOG_INFO {
		d.INFO.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warn implement Logger
func (d *DefaultLogger) Warn(v ...interface{}) {
	if d.level <= LOG_WARNING {
		d.WARN.Output(2, fmt.Sprint(v...))
	}
}

// Warnf implement Logger
func (d *DefaultLogger) Warnf(format string, v ...interface{}) {
	if d.level <= LOG_WARNING {
		d.WARN.Output(2, fmt.Sprintf(format, v...))
	}
}

// Level implement Logger
func (d *DefaultLogger) Level() LogLevel {
	return d.level
}

// SetLevel implement Logger
func (d *DefaultLogger) SetLevel(l LogLevel) {
	d.level = l
}
