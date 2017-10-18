// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"bytes"

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
