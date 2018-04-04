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

// Package javascript implements a Javascript scripting engine.
package javascript

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/scripting"
	"github.com/robertkrimen/otto"
)

type js struct {
	options scripting.Options
}

// New returns a new Javascript scripting engine.
func New(options scripting.Options) scripting.Engine {
	return &js{options}
}

// Run executes the Javascript script in the environment env and returns the output.
func (j *js) Run(ctx context.Context, script string, env map[string]interface{}) (val interface{}, err error) {
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
				err = scripting.ErrRuntime.NewWithCause(nil, val)
			default:
				err = scripting.ErrRuntime.New(nil)
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
			panic(errors.ErrContextDeadlineExceeded.New(nil))
		}
	}()

	output, err := vm.Run(script)
	if err != nil {
		return nil, scripting.ErrRuntime.NewWithCause(nil, err)
	}

	return output.Export()
}
