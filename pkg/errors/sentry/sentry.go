// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package sentry

import (
	"runtime"

	raven "github.com/getsentry/raven-go"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// ErrorAsExceptions converts the error into a raven.Exceptions.
func ErrorAsExceptions(err error, includePaths ...string) *raven.Exceptions {
	errStack := errors.Stack(err)
	exceptions := &raven.Exceptions{
		Values: make([]*raven.Exception, len(errStack)),
	}
	for i, err := range errStack {
		exception := &raven.Exception{
			Value: err.Error(),
		}
		if ttnErr, ok := errors.From(err); ok {
			exception.Value = ttnErr.MessageFormat()
			exception.Type = ttnErr.Name()
			exception.Module = ttnErr.Namespace()
			var frames []*raven.StacktraceFrame
			for _, f := range ttnErr.StackTrace() { // copied from raven-go
				pc := uintptr(f) - 1
				fn := runtime.FuncForPC(pc)
				var file string
				var line int
				if fn != nil {
					file, line = fn.FileLine(pc)
				} else {
					file = "unknown"
				}
				frame := raven.NewStacktraceFrame(pc, fn.Name(), file, line, 3, includePaths)
				if frame != nil {
					frames = append([]*raven.StacktraceFrame{frame}, frames...)
				}
			}
			exception.Stacktrace = &raven.Stacktrace{Frames: frames}
		}
		exceptions.Values[len(errStack)-i-1] = exception
	}
	return exceptions
}
