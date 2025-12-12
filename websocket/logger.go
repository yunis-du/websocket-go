package websocket

import (
	"log"
	"os"
)

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Error(v ...any)
	Errorf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Debug(v ...any)
	Debugf(format string, v ...any)
}

// EmptyLogger returns a new empty logger.
func EmptyLogger() *emptyLogger {
	return &emptyLogger{}
}

// emptyLogger is a logger that does nothing.
type emptyLogger struct{}

func (l *emptyLogger) Error(v ...any)                 {}
func (l *emptyLogger) Errorf(format string, v ...any) {}
func (l *emptyLogger) Info(v ...any)                  {}
func (l *emptyLogger) Infof(format string, v ...any)  {}
func (l *emptyLogger) Debug(v ...any)                 {}
func (l *emptyLogger) Debugf(format string, v ...any) {}

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

func (l *StdLogger) Error(v ...any) {
	l.Printf("ERROR: %s", v...)
}

func (l *StdLogger) Errorf(format string, v ...any) {
	l.Printf("ERROR: "+format, v...)
}

func (l *StdLogger) Info(v ...any) {
	l.Printf("INFO: %s", v...)
}

func (l *StdLogger) Infof(format string, v ...any) {
	l.Printf("INFO: "+format, v...)
}

func (l *StdLogger) Debug(v ...any) {
	l.Printf("DEBUG: %s", v...)
}

func (l *StdLogger) Debugf(format string, v ...any) {
	l.Printf("DEBUG: "+format, v...)
}
