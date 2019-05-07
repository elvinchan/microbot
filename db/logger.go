package db

import (
	"fmt"
	"io"
	"log"
)

// default log options
const (
	DEFAULT_LOG_PREFIX = "[microbot]"
	DEFAULT_LOG_FLAG   = log.Ldate | log.Lmicroseconds
	DEFAULT_LOG_LEVEL  = LOG_INFO
)

type DefaultLogger struct {
	DEBUG   *log.Logger
	ERR     *log.Logger
	INFO    *log.Logger
	WARN    *log.Logger
	level   LogLevel
	showSQL bool
}

var _ ILogger = &DefaultLogger{}

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

// Debug implement ILogger
func (d *DefaultLogger) Debug(v ...interface{}) {
	if d.level <= LOG_DEBUG {
		d.DEBUG.Output(2, fmt.Sprint(v...))
	}
}

// Debugf implement ILogger
func (d *DefaultLogger) Debugf(format string, v ...interface{}) {
	if d.level <= LOG_DEBUG {
		d.DEBUG.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error implement ILogger
func (d *DefaultLogger) Error(v ...interface{}) {
	if d.level <= LOG_ERR {
		d.ERR.Output(2, fmt.Sprint(v...))
	}
}

// Errorf implement ILogger
func (d *DefaultLogger) Errorf(format string, v ...interface{}) {
	if d.level <= LOG_ERR {
		d.ERR.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info implement ILogger
func (d *DefaultLogger) Info(v ...interface{}) {
	if d.level <= LOG_INFO {
		d.INFO.Output(2, fmt.Sprint(v...))
	}
}

// Infof implement ILogger
func (d *DefaultLogger) Infof(format string, v ...interface{}) {
	if d.level <= LOG_INFO {
		d.INFO.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warn implement ILogger
func (d *DefaultLogger) Warn(v ...interface{}) {
	if d.level <= LOG_WARNING {
		d.WARN.Output(2, fmt.Sprint(v...))
	}
}

// Warnf implement ILogger
func (d *DefaultLogger) Warnf(format string, v ...interface{}) {
	if d.level <= LOG_WARNING {
		d.WARN.Output(2, fmt.Sprintf(format, v...))
	}
}

// Level implement ILogger
func (d *DefaultLogger) Level() LogLevel {
	return d.level
}

// SetLevel implement ILogger
func (d *DefaultLogger) SetLevel(l LogLevel) {
	d.level = l
}

// ShowSQL implement ILogger
func (d *DefaultLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		d.showSQL = true
		return
	}
	d.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (d *DefaultLogger) IsShowSQL() bool {
	return d.showSQL
}
