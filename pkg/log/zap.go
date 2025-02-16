package log

import (
	_log "log"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ZapLogger is a zap logger implementation.
type ZapLogger struct {
	logger *zap.SugaredLogger
}

func logErrorIfNeeded(err error) {
	// https://github.com/uber-go/zap/issues/772
	// https://github.com/uber-go/zap/issues/991#issuecomment-962098428
	if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) {
		_log.Println("Sync error: ", err)
	}
}

// Errorf - log error with formatting.
func (z *ZapLogger) Errorf(format string, args ...interface{}) {
	z.logger.Errorf(format, args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Error - log error without formatting.
func (z *ZapLogger) Error(args ...interface{}) {
	z.logger.Error(args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Fatalf - log and fall with error with formatting.
func (z *ZapLogger) Fatalf(format string, args ...interface{}) {
	z.logger.Fatalf(format, args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Fatal - log and fall with error.
func (z *ZapLogger) Fatal(args ...interface{}) {
	z.logger.Fatal(args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Infof - log info messages with formatting.
func (z *ZapLogger) Infof(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Info - log info messages without formatting.
func (z *ZapLogger) Info(args ...interface{}) {
	z.logger.Info(args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Warnf - log warnings with formatting.
func (z *ZapLogger) Warnf(format string, args ...interface{}) {
	z.logger.Warnf(format, args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Warn - log warnings without formatting.
func (z *ZapLogger) Warn(args ...interface{}) {
	z.logger.Warn(args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Debugf - log debug messages with formatting.
func (z *ZapLogger) Debugf(format string, args ...interface{}) {
	z.logger.Debugf(format, args...)
	logErrorIfNeeded(z.logger.Sync())
}

// Debug - log debug messages without formatting.
func (z *ZapLogger) Debug(args ...interface{}) {
	z.logger.Debug(args...)
	logErrorIfNeeded(z.logger.Sync())
}

// NewZapLogger returns new zap logger implementation.
func NewZapLogger(isDebug bool) (*ZapLogger, error) {
	var logger *zap.Logger

	const callerSkipCount = 2

	// AddCallerSkip is needed to print real caller, not this file (zap.go).
	var err error
	if isDebug {
		logger, err = zap.NewDevelopment(zap.AddCallerSkip(callerSkipCount))
	} else {
		logger, err = zap.NewProduction(zap.AddCallerSkip(callerSkipCount))
	}

	if err != nil {
		return nil, errors.Wrap(err, "can't initialize zap logger")
	}

	return &ZapLogger{
		logger: logger.Sugar(),
	}, nil
}
