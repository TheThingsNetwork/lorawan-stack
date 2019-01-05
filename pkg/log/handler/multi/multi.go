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

// Package multi implements a pkg/log.Handler that applies every log message on multiple Handlers
package multi

import (
	"go.thethings.network/lorawan-stack/pkg/log"
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
