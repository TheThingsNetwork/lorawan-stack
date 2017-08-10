// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"fmt"
	"os"
	"time"
)

// Entry is the interface of log entries
type Entry interface {
	Level() Level
	Fields() Fielder
	Message() string
	Timestamp() time.Time
}

// entry implements Entry
type entry struct {
	logger  *Logger
	level   Level
	message string
	time    time.Time
	fields  *F
}

// interface assertions
var _ Entry = &entry{}
var _ Interface = &entry{}

// Level implements Entry
func (e *entry) Level() Level {
	return e.level
}

// Fields implements Entry
func (e *entry) Fields() Fielder {
	return e.fields
}

// Timestamp implements Entry
func (e *entry) Timestamp() time.Time {
	return e.time
}

// Message implements Entry
func (e *entry) Message() string {
	return e.message
}

// commit commits the log entry and passes it on to the handler
func (e *entry) commit(level Level, msg string) {
	// only log the entry if it has the correct level
	if e.logger.Handler != nil && e.logger.Level <= level {
		e.logger.HandleLog(&entry{
			message: msg,
			level:   level,
			time:    time.Now(),
			fields:  e.fields,
		})
	}

	// exit with fatal log level
	if level >= Fatal {
		os.Exit(1)
	}
}

// Debug implements log.Interface
func (e *entry) Debug(msg string) {
	e.commit(Debug, msg)
}

// Info implements log.Interface
func (e *entry) Info(msg string) {
	e.commit(Info, msg)
}

// Warn implements log.Interface
func (e *entry) Warn(msg string) {
	e.commit(Warn, msg)
}

// Error implements log.Interface
func (e *entry) Error(msg string) {
	e.commit(Error, msg)
}

// Fatal implements log.Interface
func (e *entry) Fatal(msg string) {
	e.commit(Fatal, msg)
}

// Debugf implements log.Interface
func (e *entry) Debugf(msg string, v ...interface{}) {
	e.Debug(fmt.Sprintf(msg, v...))
}

// Infof implements log.Interface
func (e *entry) Infof(msg string, v ...interface{}) {
	e.Info(fmt.Sprintf(msg, v...))
}

// Warnf implements log.Interface
func (e *entry) Warnf(msg string, v ...interface{}) {
	e.Warn(fmt.Sprintf(msg, v...))
}

// Errorf implements log.Interface
func (e *entry) Errorf(msg string, v ...interface{}) {
	e.Error(fmt.Sprintf(msg, v...))
}

// Fatalf implements log.Interface
func (e *entry) Fatalf(msg string, v ...interface{}) {
	e.Fatal(fmt.Sprintf(msg, v...))
}

// WithField implements log.Interface
func (e *entry) WithField(name string, value interface{}) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithField(name, value),
	}
}

// WithFields implements log.Interface
func (e *entry) WithFields(fields Fielder) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithFields(fields),
	}
}

// WithError implements log.Interface
func (e *entry) WithError(err error) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithError(err),
	}
}
