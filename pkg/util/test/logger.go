// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/test"
)

// GetLogger returns a logger for tests.
func GetLogger(t *testing.T, tag string) log.Interface {
	logger := &log.Logger{
		Level:   log.DebugLevel,
		Handler: test.NewTestingHandler(t),
	}

	return logger.WithField("tag", tag)
}
