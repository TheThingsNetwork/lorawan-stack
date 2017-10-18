// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/test"
)

// GetLogger returns a logger for tests.
func GetLogger(logger test.Logger) log.Stack {
	return &log.Logger{
		Level:   log.DebugLevel,
		Handler: test.NewTestingHandler(logger),
	}
}
