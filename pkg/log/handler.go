// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

// Handler is the interface of things that can handle log entries.
type Handler interface {
	HandleLog(Entry) error
}

// NoopHandler is a handler that does nothing.
var NoopHandler = &noopHandler{}

// noopHandler is a handler that does nothing.
type noopHandler struct{}

// HandleLog implements Handler.
func (h *noopHandler) HandleLog(Entry) error {
	return nil
}

// HandlerFunc is a function that implements Handler.
type HandlerFunc func(Entry) error

// HandleLog implements Handler.
func (fn HandlerFunc) HandleLog(e Entry) error {
	return fn(e)
}
