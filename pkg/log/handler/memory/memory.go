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

// Package memory implements a pkg/log.Handler that saves all entries in process memory
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
