// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	"time"
)

// Entry is the interface of log entries.
type Entry interface {
	Level() Level
	Fields() Fielder
	Message() string
	Timestamp() time.Time
}

// entry implements Entry.
type entry struct {
	logger interface {
		commit(e *entry)
	}
	level   Level
	message string
	time    time.Time
	fields  *F
}

// interface assertions.
var (
	_ Entry     = &entry{}
	_ Interface = &entry{}
)

// Level implements Entry.
func (e *entry) Level() Level {
	return e.level
}

// Fields implements Entry.
func (e *entry) Fields() Fielder {
	return e.fields
}

// Timestamp implements Entry.
func (e *entry) Timestamp() time.Time {
	return e.time
}

// Message implements Entry.
func (e *entry) Message() string {
	return e.message
}

// commit commits the log entry and passes it on to the handler.
func (e *entry) commit(level Level, msg string) {
	e.logger.commit(&entry{
		message: msg,
		level:   level,
		time:    time.Now(),
		fields:  e.fields,
	})
}

// Debug implements log.Interface.
func (e *entry) Debug(args ...any) {
	e.commit(DebugLevel, fmt.Sprint(args...))
}

// Info implements log.Interface.
func (e *entry) Info(args ...any) {
	e.commit(InfoLevel, fmt.Sprint(args...))
}

// Warn implements log.Interface.
func (e *entry) Warn(args ...any) {
	e.commit(WarnLevel, fmt.Sprint(args...))
}

// Error implements log.Interface.
func (e *entry) Error(args ...any) {
	e.commit(ErrorLevel, fmt.Sprint(args...))
}

// Fatal implements log.Interface.
func (e *entry) Fatal(args ...any) {
	e.commit(FatalLevel, fmt.Sprint(args...))
}

// Debugf implements log.Interface.
func (e *entry) Debugf(msg string, v ...any) {
	e.Debug(fmt.Sprintf(msg, v...))
}

// Infof implements log.Interface.
func (e *entry) Infof(msg string, v ...any) {
	e.Info(fmt.Sprintf(msg, v...))
}

// Warnf implements log.Interface.
func (e *entry) Warnf(msg string, v ...any) {
	e.Warn(fmt.Sprintf(msg, v...))
}

// Errorf implements log.Interface.
func (e *entry) Errorf(msg string, v ...any) {
	e.Error(fmt.Sprintf(msg, v...))
}

// Fatalf implements log.Interface.
func (e *entry) Fatalf(msg string, v ...any) {
	e.Fatal(fmt.Sprintf(msg, v...))
}

// WithField implements log.Interface.
func (e *entry) WithField(name string, value any) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithField(name, value),
	}
}

// WithFields implements log.Interface.
func (e *entry) WithFields(fields Fielder) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithFields(fields),
	}
}

// WithError implements log.Interface.
func (e *entry) WithError(err error) Interface {
	return &entry{
		logger:  e.logger,
		time:    e.time,
		message: e.message,
		level:   e.level,
		fields:  e.fields.WithError(err),
	}
}
