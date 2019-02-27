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

package web

import (
	"io"
	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

// NewNoopLogger returns an echo.Logger that discards all log messages.
func NewNoopLogger() echo.Logger {
	return &noopLogger{}
}

type noopLogger struct{}

func (n *noopLogger) Output() io.Writer                         { return ioutil.Discard }
func (n *noopLogger) SetOutput(w io.Writer)                     {}
func (n *noopLogger) Prefix() string                            { return "" }
func (n *noopLogger) SetPrefix(p string)                        {}
func (n *noopLogger) Level() log.Lvl                            { return log.DEBUG }
func (n *noopLogger) SetLevel(v log.Lvl)                        {}
func (n *noopLogger) SetHeader(h string)                        {}
func (n *noopLogger) Print(i ...interface{})                    {}
func (n *noopLogger) Printf(format string, args ...interface{}) {}
func (n *noopLogger) Printj(j log.JSON)                         {}
func (n *noopLogger) Debug(i ...interface{})                    {}
func (n *noopLogger) Debugf(format string, args ...interface{}) {}
func (n *noopLogger) Debugj(j log.JSON)                         {}
func (n *noopLogger) Info(i ...interface{})                     {}
func (n *noopLogger) Infof(format string, args ...interface{})  {}
func (n *noopLogger) Infoj(j log.JSON)                          {}
func (n *noopLogger) Warn(i ...interface{})                     {}
func (n *noopLogger) Warnf(format string, args ...interface{})  {}
func (n *noopLogger) Warnj(j log.JSON)                          {}
func (n *noopLogger) Error(i ...interface{})                    {}
func (n *noopLogger) Errorf(format string, args ...interface{}) {}
func (n *noopLogger) Errorj(j log.JSON)                         {}
func (n *noopLogger) Fatal(i ...interface{})                    {}
func (n *noopLogger) Fatalj(j log.JSON)                         {}
func (n *noopLogger) Fatalf(format string, args ...interface{}) {}
func (n *noopLogger) Panic(i ...interface{})                    {}
func (n *noopLogger) Panicj(j log.JSON)                         {}
func (n *noopLogger) Panicf(format string, args ...interface{}) {}
