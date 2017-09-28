// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Interface is the interface for logging TTN.
type Interface interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Debugf(msg string, v ...interface{})
	Infof(msg string, v ...interface{})
	Warnf(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Fatalf(msg string, v ...interface{})
	WithField(string, interface{}) Interface
	WithFields(Fielder) Interface
	WithError(error) Interface
}

// Stack is the interface of loggers that have a handler stack with middleware.
type Stack interface {
	// Log is an Interface.
	Interface

	// Use installs the specified middleware in the middleware stack.
	Use(Middleware)
}
