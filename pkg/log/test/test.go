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
