// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Default is the default logger used for the package global logging functions
var Default, _ = NewLogger()

// Debug calls Default.Debug
func Debug(msg string) {
	Default.Debug(msg)
}

// Info calls Default.Info
func Info(msg string) {
	Default.Info(msg)
}

// Warn calls Default.Warn
func Warn(msg string) {
	Default.Warn(msg)
}

// Error calls Default.Error
func Error(msg string) {
	Default.Error(msg)
}

// Fatal calls Default.Fatal
func Fatal(msg string) {
	Default.Fatal(msg)
}

// Debugf calls Default.Debugf
func Debugf(msg string, v ...interface{}) {
	Default.Debugf(msg, v...)
}

// Infof calls Default.Infof
func Infof(msg string, v ...interface{}) {
	Default.Infof(msg, v...)
}

// Warnf calls Default.Warnf
func Warnf(msg string, v ...interface{}) {
	Default.Warnf(msg, v...)
}

// Errorf calls Default.Errorf
func Errorf(msg string, v ...interface{}) {
	Default.Errorf(msg, v...)
}

// Fatalf calls Default.Fatalf
func Fatalf(msg string, v ...interface{}) {
	Default.Fatalf(msg, v...)
}

// WithField calls Default.WithField
func WithField(k string, v interface{}) Interface {
	return Default.WithField(k, v)
}

// WithFields calls Default.WithFields
func WithFields(f Fielder) Interface {
	return Default.WithFields(f)
}

// WithError calls Default.WithError
func WithError(err error) Interface {
	return Default.WithError(err)
}
