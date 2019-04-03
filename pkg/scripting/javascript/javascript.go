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

	"github.com/robertkrimen/otto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/scripting"
)

type js struct {
	options scripting.Options
}

// New returns a new Javascript scripting engine.
func New(options scripting.Options) scripting.Engine {
	return &js{options}
}

var errRuntime = errors.Define("runtime", "runtime error")

// Run executes the Javascript script in the environment env and returns the output.
func (j *js) Run(ctx context.Context, script string, env map[string]interface{}) (val interface{}, err error) {
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

	vm := otto.New()
	vm.SetStackDepthLimit(j.options.StackDepthLimit)

	err = vm.Set("env", env)
	if err != nil {
		return
	}

	defer func() {
		if caught := recover(); caught != nil {
			switch val := caught.(type) {
			case error:
				err = errRuntime.WithCause(val)
			default:
				err = errRuntime
			}
			return
		}
	}()

	vm.Interrupt = make(chan func(), 1)
	ctx, cancel := context.WithTimeout(ctx, j.options.Timeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		vm.Interrupt <- func() {
			panic(context.DeadlineExceeded)
		}
	}()

	output, err := vm.Run(script)
	if err != nil {
		return nil, errRuntime.WithCause(err)
	}

	return output.Export()
}
