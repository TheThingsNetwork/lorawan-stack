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
	"os"
	"sync"
)

var defaultOptions = []Option{
	WithLevel(InfoLevel),
	WithHandler(NewCLI(os.Stdout)),
}

// NewLogger creates a new logger with the default options.
func NewLogger(opts ...Option) (*Logger, error) {
	logger := &Logger{}

	for _, opt := range append(defaultOptions, opts...) {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}

	return logger, nil
}

// Logger implements Stack.
type Logger struct {
	mutex      sync.RWMutex
	Level      Level
	Handler    Handler
	middleware []Middleware
	stack      Handler
}

// Use installs the handler middleware.
func (l *Logger) Use(middleware Middleware) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.middleware = append(l.middleware, middleware)

	// the first handler uses the base handler from the logger
	// we need to wrap this is a function so the base handler can be
	// changed afterwards.
	handler := Handler(HandlerFunc(func(entry Entry) error {
		return l.Handler.HandleLog(entry)
	}))

	for i := len(l.middleware) - 1; i >= 0; i-- {
		handler = l.middleware[i].Wrap(handler)
	}

	l.stack = handler
}

// commit comits the entry to the handler.
func (l *Logger) commit(e *entry) {
	handler := l.stack
	if handler == nil {
		handler = l.Handler
	}

	if handler != nil && l.Level <= e.level {
		l.mutex.RLock()
		defer l.mutex.RUnlock()
		_ = handler.HandleLog(e)
	}

	if e.Level() == FatalLevel {
		os.Exit(1)
	}
}

// Debug implements log.Interface.
func (l *Logger) Debug(msg string) {
	l.entry().commit(DebugLevel, msg)
}

// Info implements log.Interface.
func (l *Logger) Info(msg string) {
	l.entry().commit(InfoLevel, msg)
}

// Warn implements log.Interface.
func (l *Logger) Warn(msg string) {
	l.entry().commit(WarnLevel, msg)
}

// Error implements log.Interface.
func (l *Logger) Error(msg string) {
	l.entry().commit(ErrorLevel, msg)
}

// Fatal implements log.Interface.
func (l *Logger) Fatal(msg string) {
	l.entry().commit(FatalLevel, msg)
}

// Debugf implements log.Interface.
func (l *Logger) Debugf(msg string, v ...interface{}) {
	l.Debug(fmt.Sprintf(msg, v...))
}

// Infof implements log.Interface.
func (l *Logger) Infof(msg string, v ...interface{}) {
	l.Info(fmt.Sprintf(msg, v...))
}

// Warnf implements log.Interface.
func (l *Logger) Warnf(msg string, v ...interface{}) {
	l.Warn(fmt.Sprintf(msg, v...))
}

// Errorf implements log.Interface.
func (l *Logger) Errorf(msg string, v ...interface{}) {
	l.Error(fmt.Sprintf(msg, v...))
}

// Fatalf implements log.Interface.
func (l *Logger) Fatalf(msg string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(msg, v...))
}

// WithField implements log.Interface.
func (l *Logger) WithField(name string, val interface{}) Interface {
	return l.entry().WithField(name, val)
}

// WithFields implements log.Interface.
func (l *Logger) WithFields(fields Fielder) Interface {
	return l.entry().WithFields(fields)
}

// WithError implements log.Interface.
func (l *Logger) WithError(err error) Interface {
	return l.entry().WithError(err)
}

// entry creates a new log entry.
func (l *Logger) entry() *entry {
	return &entry{
		logger: l,
		fields: Fields(),
	}
}
