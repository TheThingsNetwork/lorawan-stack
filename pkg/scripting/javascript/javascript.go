// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
