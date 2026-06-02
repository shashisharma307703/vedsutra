package logger

import (
	"log"
	"os"
)

// Logger defines the logging interface used across the application
type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// CLILogger is a simple command-line logger implementation
type CLILogger struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
}

// NewCLILogger creates a new CLI logger instance
func NewCLILogger() *CLILogger {
	return &CLILogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

// Info logs an info level message
func (l *CLILogger) Info(msg string, args ...interface{}) {
	l.infoLogger.Printf(msg, args...)
}

// Debug logs a debug level message
func (l *CLILogger) Debug(msg string, args ...interface{}) {
	l.debugLogger.Printf(msg, args...)
}

// Error logs an error level message
func (l *CLILogger) Error(msg string, args ...interface{}) {
	l.errorLogger.Printf(msg, args...)
}
