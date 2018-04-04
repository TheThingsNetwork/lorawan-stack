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

package test

import (
	"bytes"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Logger represents the logging interface implemented by i.e. testing.T
type Logger interface {
	Fatal(args ...interface{})
	Log(args ...interface{})
}

// TestingHandler implements Handler.
type TestingHandler struct {
	logger     Logger
	cliHandler *log.CLIHandler
	bufferMu   sync.Mutex
	buffer     *bytes.Buffer
}

// NewTestingHandler returns a new TestingHandler.
func NewTestingHandler(l Logger) *TestingHandler {
	buffer := bytes.NewBuffer([]byte{})

	return &TestingHandler{
		logger:     l,
		cliHandler: log.NewCLI(buffer),
		buffer:     buffer,
	}
}

// HandleLog implements Handler.
func (h *TestingHandler) HandleLog(e log.Entry) error {
	h.bufferMu.Lock()
	defer h.bufferMu.Unlock()

	defer h.buffer.Reset()
	if err := h.cliHandler.HandleLog(e); err != nil {
		return err
	}

	if e.Level() == log.FatalLevel {
		h.logger.Fatal(h.buffer.String())
	} else {
		h.logger.Log(h.buffer.String())
	}

	return nil
}
