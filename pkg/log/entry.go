// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"fmt"
	"time"
)

// Entry is the interface of log entries
type Entry interface {
	Level() Level
	Fields() Fielder
	Message() string
	Timestamp() time.Time
}

type E struct {
	level   Level
	message string
	time    time.Time
	fields  *F
	logger  *Logger
}

var _ Entry = &E{}
var _ Interface = &E{}

// Level implements Entry
func (e *E) Level() Level {
	return e.level
}

// Fields implements Entry
func (e *E) Fields() Fielder {
	return e.fields
}

// Timestamp implements Entry
func (e *E) Timestamp() time.Time {
	return e.time
}

// Message implements Entry
func (e *E) Message() string {
	return e.message
}

func (e *E) commit(level Level, msg string) {
	e.logger.Handler.HandleLog(e)
}

// Debug implements log.Interface
func (e *E) Debug(msg string) {
	e.commit(Debug, msg)
}

// Info implements log.Interface
func (e *E) Info(msg string) {
	e.commit(Info, msg)
}

// Warn implements log.Interface
func (e *E) Warn(msg string) {
	e.commit(Warn, msg)
}

// Error implements log.Interface
func (e *E) Error(msg string) {
	e.commit(Error, msg)
}

// Fatal implements log.Interface
func (e *E) Fatal(msg string) {
	e.commit(Fatal, msg)
}

// Debugf implements log.Interface
func (e *E) Debugf(msg string, v ...interface{}) {
	e.Debug(fmt.Sprintf(msg, v...))
}

// Infof implements log.Interface
func (e *E) Infof(msg string, v ...interface{}) {
	e.Info(fmt.Sprintf(msg, v...))
}

// Warnf implements log.Interface
func (e *E) Warnf(msg string, v ...interface{}) {
	e.Warn(fmt.Sprintf(msg, v...))
}

// Errorf implements log.Interface
func (e *E) Errorf(msg string, v ...interface{}) {
	e.Error(fmt.Sprintf(msg, v...))
}

// Fatalf implements log.Interface
func (e *E) Fatalf(msg string, v ...interface{}) {
	e.Fatal(fmt.Sprintf(msg, v...))
}

// WithField implements log.Interface
func (e *E) WithField(name string, value interface{}) Interface {
	return &E{
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithField(name, value),
	}
}

// WithFields implements log.Interface
func (e *E) WithFields(fields Fielder) Interface {
	return &E{
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithFields(fields),
	}
}

// WithError implements log.Interface
func (e *E) WithError(err error) Interface {
	return &E{
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithError(err),
	}
}
