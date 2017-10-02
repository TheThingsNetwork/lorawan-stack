// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sentry

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	raven "github.com/getsentry/raven-go"
)

// Sentry is a log.Handler that sends errors to Sentry.
type Sentry struct {
	*raven.Client
}

// New creates a new Sentry log middleware.
func New(client *raven.Client) log.Middleware {
	if client == nil {
		client = raven.DefaultClient
	}
	return &Sentry{Client: client}
}

// NewWithDSN returns a new Sentry log middleware that uses the given DSN
func NewWithDSN(dsn string) log.Middleware {
	client, _ := raven.New(dsn)
	return New(client)
}

// Wrap an existing log handler with Sentry.
func (s *Sentry) Wrap(next log.Handler) log.Handler {
	return log.HandlerFunc(func(entry log.Entry) (err error) {
		switch entry.Level() {
		case log.ErrorLevel:
			s.forward(entry, false)
		case log.FatalLevel:
			s.forward(entry, true)
		}
		err = next.HandleLog(entry)
		return
	})
}

func (s *Sentry) forward(e log.Entry, wait bool) {
	fields := e.Fields().Fields()
	var err error
	if fieldsErr, ok := fields["error"]; ok {
		if fieldsErr, ok := fieldsErr.(error); ok {
			err = fieldsErr
		}
	}
	details := make(map[string]string)
	if err == nil {
		err = errors.New(e.Message())
	} else {
		details["log_message"] = e.Message()
	}
	for k, v := range fields {
		if k != "error" {
			details[k] = fmt.Sprint(v)
		}
	}
	details["log_level"] = e.Level().String()
	trace := raven.NewStacktrace(6, 3, []string{"github.com/TheThings"})
	if wait {
		s.Client.CaptureMessageAndWait(err.Error(), details, trace)
	} else {
		s.Client.CaptureMessage(err.Error(), details, trace) // non-blocking
	}
}
