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

// Package log contains all the structs to log TTN.
package log

// Interface is the interface for logging TTN.
type Interface interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Debugf(msg string, v ...any)
	Infof(msg string, v ...any)
	Warnf(msg string, v ...any)
	Errorf(msg string, v ...any)
	Fatalf(msg string, v ...any)
	WithField(string, any) Interface
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
