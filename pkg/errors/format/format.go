// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package format

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

type errorLogger struct {
	*log.Logger
}

func (e errorLogger) WithError(err error, msg string) {
	e.Logger.WithError(err).Error(msg)
}

func init() {
	errors.FormatErrorSignaler = errorLogger{Logger: log.Default}
}
