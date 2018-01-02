// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package memory

import (
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Handler implements log.Handler by storing entries in memory.
type Handler struct {
	mu      sync.Mutex
	Entries []log.Entry
}

// New creates a new Handler that stores the entries in memory.
func New() *Handler {
	return &Handler{
		Entries: make([]log.Entry, 0),
	}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(entry log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Entries = append(h.Entries, entry)

	return nil
}
