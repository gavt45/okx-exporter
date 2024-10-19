// Package log implements logger api.
package log

import (
	"log"
)

// Logger instance.
var Logger ILogger

// ILogger - logger infterface.
type ILogger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

// SetLogger sets logger.
func SetLogger(logger ILogger) {
	Logger = logger
}

func initIfNull() {
	if Logger == nil {
		zlog, err := NewZapLogger(true)
		if err != nil {
			log.Fatal(err)
			return
		}
		SetLogger(zlog)
	}
}

// Errorf - log error with formatting.
func Errorf(format string, args ...interface{}) {
	initIfNull()
	Logger.Errorf(format, args...)
}

// Error - log error without formatting.
func Error(args ...interface{}) {
	initIfNull()
	Logger.Error(args...)
}

// Fatalf - log and fall with error with formatting.
func Fatalf(format string, args ...interface{}) {
	initIfNull()
	Logger.Fatalf(format, args...)
}

// Fatal - log and fall with error.
func Fatal(args ...interface{}) {
	initIfNull()
	Logger.Fatal(args...)
}

// Infof - log info messages with formatting.
func Infof(format string, args ...interface{}) {
	initIfNull()
	Logger.Infof(format, args...)
}

// Info - log info messages without formatting.
func Info(args ...interface{}) {
	initIfNull()
	Logger.Info(args...)
}

// Warnf - log warnings with formatting.
func Warnf(format string, args ...interface{}) {
	initIfNull()
	Logger.Warnf(format, args...)
}

// Warn - log warnings without formatting.
func Warn(args ...interface{}) {
	initIfNull()
	Logger.Warn(args...)
}

// Debugf - log debug messages with formatting.
func Debugf(format string, args ...interface{}) {
	initIfNull()
	Logger.Debugf(format, args...)
}

// Debug - log debug messages without formatting.
func Debug(args ...interface{}) {
	initIfNull()
	Logger.Debug(args...)
}
