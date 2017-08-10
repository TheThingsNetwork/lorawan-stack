// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"fmt"
)

var defaultOptions = []Option{
	WithLevel(Info),
	WithHandler(NoopHandler),
}

// NewLogger creates a new logger with the default options
func NewLogger(opts ...Option) (Interface, error) {
	logger := &Logger{}

	for _, opt := range append(defaultOptions, opts...) {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}

	return logger, nil
}

// Logger implements Interface
type Logger struct {
	Handler
	Level Level
}

// Debug implements log.Interface
func (l *Logger) Debug(msg string) {
	l.entry().commit(Debug, msg)
}

// Info implements log.Interface
func (l *Logger) Info(msg string) {
	l.entry().commit(Info, msg)
}

// Warn implements log.Interface
func (l *Logger) Warn(msg string) {
	l.entry().commit(Warn, msg)
}

// Error implements log.Interface
func (l *Logger) Error(msg string) {
	l.entry().commit(Error, msg)
}

// Fatal implements log.Interface
func (l *Logger) Fatal(msg string) {
	l.entry().commit(Fatal, msg)
}

// Debugf implements log.Interface
func (l *Logger) Debugf(msg string, v ...interface{}) {
	l.Debug(fmt.Sprintf(msg, v...))
}

// Infof implements log.Interface
func (l *Logger) Infof(msg string, v ...interface{}) {
	l.Info(fmt.Sprintf(msg, v...))
}

// Warnf implements log.Interface
func (l *Logger) Warnf(msg string, v ...interface{}) {
	l.Warn(fmt.Sprintf(msg, v...))
}

// Errorf implements log.Interface
func (l *Logger) Errorf(msg string, v ...interface{}) {
	l.Error(fmt.Sprintf(msg, v...))
}

// Fatalf implements log.Interface
func (l *Logger) Fatalf(msg string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(msg, v...))
}

// WithField implements log.Interface
func (l *Logger) WithField(name string, val interface{}) Interface {
	return l.entry().WithField(name, val)
}

// WithFields implements log.Interface
func (l *Logger) WithFields(fields Fielder) Interface {
	return l.entry().WithFields(fields)
}

// WithError implements log.Interface
func (l *Logger) WithError(err error) Interface {
	return l.entry().WithError(err)
}

// entry creates a new log entry
func (l *Logger) entry() *entry {
	return &entry{
		logger: l,
		fields: Fields(),
	}
}
