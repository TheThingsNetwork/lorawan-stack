// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package web

import (
	"io"
	"io/ioutil"

	"github.com/labstack/gommon/log"
)

type noopLogger struct{}

func (n *noopLogger) Output() io.Writer                         { return ioutil.Discard }
func (n *noopLogger) SetOutput(w io.Writer)                     {}
func (n *noopLogger) Prefix() string                            { return "" }
func (n *noopLogger) SetPrefix(p string)                        {}
func (n *noopLogger) Level() log.Lvl                            { return log.DEBUG }
func (n *noopLogger) SetLevel(v log.Lvl)                        {}
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
