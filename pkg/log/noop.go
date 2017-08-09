// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Noop just does nothing
var Noop = &noop{}

// noop is a log.Interface that does nothing
type noop struct{}

// Debug implements log.Interface
func (n noop) Debug(msg string) {}

// Info implements log.Interface
func (n noop) Info(msg string) {}

// Warn implements log.Interface
func (n noop) Warn(msg string) {}

// Error implements log.Interface
func (n noop) Error(msg string) {}

// Fatal implements log.Interface
func (n noop) Fatal(msg string) {}

// Debugf implements log.Interface
func (n noop) Debugf(msg string, v ...interface{}) {}

// Infof implements log.Interface
func (n noop) Infof(msg string, v ...interface{}) {}

// Warnf implements log.Interface
func (n noop) Warnf(msg string, v ...interface{}) {}

// Errorf implements log.Interface
func (n noop) Errorf(msg string, v ...interface{}) {}

// Fatalf implements log.Interface
func (n noop) Fatalf(msg string, v ...interface{}) {}

// WithField implements log.Interface
func (n noop) WithField(string, interface{}) Interface { return n }

// WithFields implements log.Interface
func (n noop) WithFields(Fielder) Interface { return n }

// WithError implements log.Interface
func (n noop) WithError(error) Interface { return n }
