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

// Package javascript implements a Javascript scripting engine.
package javascript

import (
	"context"
	"runtime/trace"
	"time"

	"github.com/dop251/goja"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/scripting"
)

type js struct {
	options scripting.Options
}

// New returns a new Javascript scripting engine.
func New(options scripting.Options) scripting.Engine {
	return &js{options}
}

var (
	errScriptTimeout      = errors.DefineDeadlineExceeded("script_timeout", "script timeout")
	errScriptInterrupt    = errors.DefineAborted("script_interrupt", "script interrupt")
	errScript             = errors.Define("script", "{message}")
	errNoScriptOutput     = errors.DefineAborted("no_script_output", "no script output")
	errRuntime            = errors.Define("runtime", "runtime error")
	errEntrypointNotFound = errors.DefineNotFound("entrypoint_not_found", "entrypoint `{entrypoint}` not found")
)

func convertError(err error) error {
	if err == nil {
		return nil
	}
	switch gojaErr := err.(type) {
	case *goja.InterruptedError:
		if gojaErr.Value() == context.DeadlineExceeded {
			return errScriptTimeout.WithCause(err)
		}
		return errScriptInterrupt.WithCause(err)
	case *goja.Exception:
		return errScript.WithAttributes("message", gojaErr.Error()).WithCause(err)
	default:
		return errRuntime.WithCause(err)
	}
}

// Run executes the Javascript script in the environment env and returns the output.
func (j *js) Run(ctx context.Context, script, fn string, params ...interface{}) (as func(target interface{}) error, err error) {
	defer trace.StartRegion(ctx, "run javascript").End()

	start := time.Now()
	defer func() {
		runLatency.Observe(time.Since(start).Seconds())
		if err != nil {
			runs.WithLabelValues("error").Inc()
		} else {
			runs.WithLabelValues("ok").Inc()
		}
	}()

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	// TODO: Set memory limit (https://github.com/dop251/goja/issues/6)

	interrupt := time.AfterFunc(j.options.Timeout, func() {
		vm.Interrupt(context.DeadlineExceeded)
	})
	defer interrupt.Stop()

	defer func() {
		if caught := recover(); caught != nil {
			switch val := caught.(type) {
			case error:
				err = errRuntime.WithCause(val)
			default:
				err = errRuntime.New()
			}
		}
	}()

	_, err = vm.RunString(script)
	if err != nil {
		return nil, convertError(err)
	}

	entrypoint, ok := goja.AssertFunction(vm.Get(fn))
	if !ok {
		return nil, errEntrypointNotFound.WithAttributes("entrypoint", fn)
	}

	args := make([]goja.Value, len(params))
	for i, param := range params {
		args[i] = vm.ToValue(param)
	}
	res, err := entrypoint(goja.Undefined(), args...)
	if err != nil {
		return nil, convertError(err)
	}

	return func(target interface{}) (err error) {
		defer func() {
			if caught := recover(); caught != nil {
				switch val := caught.(type) {
				case error:
					err = errRuntime.WithCause(val)
				default:
					err = errRuntime.New()
				}
			}
		}()
		if res == goja.Null() || res == goja.Undefined() {
			return errNoScriptOutput.New()
		}
		return convertError(vm.ExportTo(res, target))
	}, nil
}
