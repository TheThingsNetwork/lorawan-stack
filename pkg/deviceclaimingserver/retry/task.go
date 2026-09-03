// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

// Package retry provides a task that retries an operation with a backoff.
package retry

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
)

var errOperationUnsuccessful = errors.DefineInternal(
	"operation_unsuccessful",
	"operation `{name}` unsuccessful after `{number}` attempts",
)

// Task retries a function with a backoff.
type Task struct {
	Name        string
	F           func() (bool, error)
	WaitTime    time.Duration
	MaxAttempts int
	Jitter      float64
}

// Do runs the task function until one of the conditions are met.
func (t Task) Do(ctx context.Context) error {
	for count := 0; ; count++ {
		retry, err := t.F()
		switch {
		case !retry:
			return err
		case retry && count >= t.MaxAttempts:
			if err != nil {
				return errOperationUnsuccessful.WithAttributes("name", t.Name, "number", t.MaxAttempts).WithCause(err)
			}
			return errOperationUnsuccessful.WithAttributes("name", t.Name, "number", t.MaxAttempts)
		default:
			// Retry. This is just for completeness.
		}
		waitTime := random.Jitter(t.WaitTime, t.Jitter)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}
