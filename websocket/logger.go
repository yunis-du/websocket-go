package websocket

import (
	"log"
	"os"
)

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

// StdLogger is a simple implementation of Logger using the standard log package.
type StdLogger struct {
	*log.Logger
}

// NewStdLogger returns a new StdLogger.
func NewStdLogger() *StdLogger {
	return &StdLogger{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (l *StdLogger) Error(v ...interface{}) {
	l.Printf("ERROR: %s", v...)
}

func (l *StdLogger) Errorf(format string, v ...interface{}) {
	l.Printf("ERROR: "+format, v...)
}

func (l *StdLogger) Info(v ...interface{}) {
	l.Printf("INFO: %s", v...)
}

func (l *StdLogger) Infof(format string, v ...interface{}) {
	l.Printf("INFO: "+format, v...)
}

func (l *StdLogger) Debug(v ...interface{}) {
	l.Printf("DEBUG: %s", v...)
}

func (l *StdLogger) Debugf(format string, v ...interface{}) {
	l.Printf("DEBUG: "+format, v...)
}
