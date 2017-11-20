// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
