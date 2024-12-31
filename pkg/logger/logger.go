// pkg/logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger wraps the standard log package with some additional context
type Logger struct {
	component string
	logger    *log.Logger
}

// New creates a new logger for a specific component
func New(component string) *Logger {
	// Format: timestamp | component | level | message
	return &Logger{
		component: component,
		logger:    log.New(os.Stdout, "", 0),
	}
}

// formatMessage creates our consistent log format
func (l *Logger) formatMessage(level string, msg string) string {
	return fmt.Sprintf("%s | %s | %s | %s",
		time.Now().Format(time.RFC3339),
		l.component,
		level,
		msg,
	)
}

// Printf is for general logging with formatting
func (l *Logger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatMessage("INFO", msg))
}

// Print logs a message
func (l *Logger) Print(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logger.Print(l.formatMessage("INFO", msg))
}

// Println logs a message with a newline
func (l *Logger) Println(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logger.Println(l.formatMessage("INFO", msg))
}

// Fatal logs a message and calls os.Exit(1)
func (l *Logger) Fatal(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logger.Fatal(l.formatMessage("FATAL", msg))
}

// Fatalf logs a formatted message and calls os.Exit(1)
func (l *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Fatal(l.formatMessage("FATAL", msg))
}

// Panic logs a message and calls panic()
func (l *Logger) Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logger.Panic(l.formatMessage("PANIC", msg))
}

// Panicf logs a formatted message and calls panic()
func (l *Logger) Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Panic(l.formatMessage("PANIC", msg))
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Print(l.formatMessage("INFO", msg))
}

// Error logs an error message with an error
func (l *Logger) Error(msg string, err error) {
	l.logger.Print(l.formatMessage("ERROR", fmt.Sprintf("%s: %v", msg, err)))
}
