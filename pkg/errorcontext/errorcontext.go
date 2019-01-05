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

// Package errorcontext implements a Context that can be canceled with a specific error instead of the usual context.Canceled.
package errorcontext

import (
	"context"
	"sync"
)

// CancelFunc extends the regular context.CancelFunc with an error argument
type CancelFunc func(error)

// ErrorContext can be used to attach errors to a context cancelation.
type ErrorContext struct {
	context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	err    error
}

// Cancel the ErrorContext with an error. The context can not be re-used or re-canceled after this function is called.
func (c *ErrorContext) Cancel(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return
	}
	c.err = err
	c.cancel()
}

// Err returns the err of the ErrorContext if non-nil, or the error of the underlying Context.
func (c *ErrorContext) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}
	return c.Context.Err()
}

// New returns a new ErrorContext that extends the parent context
func New(parent context.Context) (context.Context, CancelFunc) {
	parent, cancel := context.WithCancel(parent)
	ctx := &ErrorContext{Context: parent, cancel: cancel}
	return ctx, ctx.Cancel
}
