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
