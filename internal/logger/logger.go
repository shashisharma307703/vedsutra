package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger defines the logging interface used across the application
type Logger interface {
	Info(msg string, args ...interface{})
	Infof(format string, args ...interface{})
	Debug(msg string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warn(msg string, args ...interface{})
	Warnf(format string, args ...interface{})
	Error(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// CLILogger is a simple command-line logger implementation
type CLILogger struct {
	name string
	infoLogger  *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger creates a new named logger instance
func NewLogger(name string) *CLILogger {
	return &CLILogger{
		name:        name,
		infoLogger:  log.New(os.Stdout, fmt.Sprintf("[%s] INFO: ", name), log.LstdFlags),
		debugLogger: log.New(os.Stdout, fmt.Sprintf("[%s] DEBUG: ", name), log.LstdFlags),
		warnLogger:  log.New(os.Stdout, fmt.Sprintf("[%s] WARN: ", name), log.LstdFlags),
		errorLogger: log.New(os.Stderr, fmt.Sprintf("[%s] ERROR: ", name), log.LstdFlags),
	}
}

// NewCLILogger creates a new CLI logger instance (deprecated, use NewLogger)
func NewCLILogger() *CLILogger {
	return NewLogger("app")
}

// Info logs an info level message
func (l *CLILogger) Info(msg string, args ...interface{}) {
	l.infoLogger.Printf(msg, args...)
}

// Infof logs an info level formatted message
func (l *CLILogger) Infof(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

// Debug logs a debug level message
func (l *CLILogger) Debug(msg string, args ...interface{}) {
	l.debugLogger.Printf(msg, args...)
}

// Debugf logs a debug level formatted message
func (l *CLILogger) Debugf(format string, args ...interface{}) {
	l.debugLogger.Printf(format, args...)
}

// Warn logs a warn level message
func (l *CLILogger) Warn(msg string, args ...interface{}) {
	l.warnLogger.Printf(msg, args...)
}

// Warnf logs a warn level formatted message
func (l *CLILogger) Warnf(format string, args ...interface{}) {
	l.warnLogger.Printf(format, args...)
}

// Error logs an error level message
func (l *CLILogger) Error(msg string, args ...interface{}) {
	l.errorLogger.Printf(msg, args...)
}

// Errorf logs an error level formatted message
func (l *CLILogger) Errorf(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}
