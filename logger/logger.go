package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// Log levels
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
)

var levelNames = map[LogLevel]string{
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
}

// Logger handles structured logging for the bot
type Logger struct {
	out       io.Writer
	level     LogLevel
	useColors bool
}

// New creates a new logger instance
func New(minLevel LogLevel, useColors bool) *Logger {
	return &Logger{
		out:       os.Stdout,
		level:     minLevel,
		useColors: useColors,
	}
}

// SetOutput sets the output destination for logs
func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(Debug, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(Info, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(Warn, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(Error, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.log(Fatal, msg, fields...)
	os.Exit(1)
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, msg string, fields ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	var output string
	if len(fields) > 0 {
		fieldsStr := fmt.Sprint(fields...)
		output = fmt.Sprintf("[%s] %s - %s | %s\n", timestamp, levelName, msg, fieldsStr)
	} else {
		output = fmt.Sprintf("[%s] %s - %s\n", timestamp, levelName, msg)
	}

	fmt.Fprint(l.out, output)
}

// WithFields allows logging with structured key-value pairs
func (l *Logger) WithFields(level LogLevel, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]

	fieldsStr := ""
	for k, v := range fields {
		fieldsStr += fmt.Sprintf("%s=%v ", k, v)
	}

	output := fmt.Sprintf("[%s] %s - %s | %s\n", timestamp, levelName, msg, fieldsStr)
	fmt.Fprint(l.out, output)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}
