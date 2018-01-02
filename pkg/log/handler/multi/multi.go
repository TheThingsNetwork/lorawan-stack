// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package multi

import (
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Handler implements log.Handler.
type Handler struct {
	handlers []log.Handler
}

// New returns a new handler that combines the underlying handlers.
func New(handlers ...log.Handler) *Handler {
	return &Handler{
		handlers: handlers,
	}
}

// HandleLog implements log.Handler.
func (m *Handler) HandleLog(entry log.Entry) error {
	var err error
	for _, handler := range m.handlers {
		e := handler.HandleLog(entry)
		// save the last error but continue
		if e != nil {
			err = e
		}
	}

	return err
}
