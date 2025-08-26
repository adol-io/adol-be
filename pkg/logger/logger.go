package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface defines logging methods
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// logrusLogger implements Logger interface using logrus
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	logger := logrus.New()
	
	// Set output to stdout
	logger.SetOutput(os.Stdout)
	
	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// Set log level from environment or default to info
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	
	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

// Debug logs a debug message
func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Info logs an info message
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn logs a warning message
func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error logs an error message
func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal logs a fatal message and exits
func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// WithField adds a field to the logger
func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields adds fields to the logger
func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
	}
}