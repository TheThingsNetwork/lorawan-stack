// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Middleware is the interface of middleware for handlers.
// It's similar to middleware in HTTP stacks, where the middleware gets access to
// the entry being logged and the next handler in the stack.
type Middleware interface {
	// Wrap decorates the next handler by doing some extra work.
	Wrap(next Handler) Handler
}

// MiddlewareFunc is a function that implements Middleware.
type MiddlewareFunc func(next Handler) Handler

// Wrap implements Middleware.
func (fn MiddlewareFunc) Wrap(next Handler) Handler {
	return fn(next)
}
