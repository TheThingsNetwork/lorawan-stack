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
func New(options scripting.Options) scripting.AheadOfTimeEngine {
	return &js{options}
}

var (
	errScriptTimeout      = errors.DefineDeadlineExceeded("script_timeout", "script timeout")
	errScriptInterrupt    = errors.DefineAborted("script_interrupt", "script interrupt")
	errScript             = errors.DefineAborted("script", "{message}")
	errNoScriptOutput     = errors.DefineAborted("no_script_output", "no script output")
	errRuntime            = errors.DefineAborted("runtime", "{message}")
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
		return errRuntime.WithAttributes("message", err.Error()).WithCause(err)
	}
}

// Run executes the Javascript script and returns the output.
func (j *js) Run(ctx context.Context, script, fn string, params ...any) (as func(target any) error, err error) {
	run := func(vm *goja.Runtime) (goja.Value, error) {
		return vm.RunString(script)
	}
	return j.run(ctx, run, fn, params...)
}

// Compile compiles the Javascript script and returns the compiled program.
func (j *js) Compile(ctx context.Context, script string) (run func(context.Context, string, ...any) (func(any) error, error), err error) {
	defer trace.StartRegion(ctx, "compile javascript").End()

	start := time.Now()
	defer func() {
		compilationsLatency.Observe(time.Since(start).Seconds())
		if err != nil {
			compilations.WithLabelValues("error").Inc()
		} else {
			compilations.WithLabelValues("ok").Inc()
		}
	}()

	program, err := goja.Compile("", script, false)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, fn string, params ...any) (func(any) error, error) {
		run := func(vm *goja.Runtime) (goja.Value, error) {
			return vm.RunProgram(program)
		}
		return j.run(ctx, run, fn, params...)
	}, nil
}

func (j *js) run(ctx context.Context, f func(*goja.Runtime) (goja.Value, error), fn string, params ...any) (as func(target any) error, err error) {
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
				err = errRuntime.WithAttributes("message", val.Error()).WithCause(val)
			default:
				err = errRuntime.WithAttributes("message", "runtime error")
			}
		}
	}()

	_, err = f(vm)
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

	return func(target any) (err error) {
		defer func() {
			if caught := recover(); caught != nil {
				switch val := caught.(type) {
				case error:
					err = errRuntime.WithAttributes("message", val.Error()).WithCause(val)
				default:
					err = errRuntime.WithAttributes("message", "runtime error")
				}
			}
		}()
		if res == goja.Null() || res == goja.Undefined() {
			return errNoScriptOutput.New()
		}
		return convertError(vm.ExportTo(res, target))
	}, nil
}
