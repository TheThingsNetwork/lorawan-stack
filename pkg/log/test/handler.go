// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"bytes"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// TestingHandler implements Handler.
type TestingHandler struct {
	tb         testing.TB
	cliHandler *log.CLIHandler
	buffer     *bytes.Buffer
}

// NewTestingHandler returns a new TestingHandler.
func NewTestingHandler(tb testing.TB) *TestingHandler {
	buffer := bytes.NewBuffer([]byte{})

	return &TestingHandler{
		tb:         tb,
		cliHandler: log.NewCLI(buffer),
		buffer:     buffer,
	}
}

// HandleLog implements Handler.
func (h *TestingHandler) HandleLog(e log.Entry) error {
	defer h.buffer.Reset()
	if err := h.cliHandler.HandleLog(e); err != nil {
		return err
	}

	if e.Level() == log.FatalLevel {
		h.tb.Fatal(h.buffer.String())
	} else {
		h.tb.Log(h.buffer.String())
	}

	return nil
}
