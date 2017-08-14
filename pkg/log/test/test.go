// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package test exposes simple implementations of log interfaces that can be used for testing.
package test

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Entry is a simple log.Entry that can be used for testing.
type Entry struct {
	M string
	L log.Level
	T time.Time
	F log.Fielder
}

// Message implements log.Entry.
func (e *Entry) Message() string {
	return e.M
}

// Level implements log.Entry.
func (e *Entry) Level() log.Level {
	return e.L
}

// Timestamp implements log.Entry.
func (e *Entry) Timestamp() time.Time {
	return e.T
}

// Fields implements log.Entry.
func (e *Entry) Fields() log.Fielder {
	return e.F
}
