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

package io

import (
	"sync/atomic"
	"time"
)

const downlinkTokenItems = 1 << 4

type downlinkToken struct {
	key            uint16
	correlationIDs []string
	time           time.Time
}

// DownlinkTokens stores a set of downlink tokens and can be used to track roundtrip time.
// The number of downlink tokens stored is fixed to 16. New issued tokens with `Next` overwrite the oldest token.
type DownlinkTokens struct {
	last  uint32
	items [downlinkTokenItems]downlinkToken
}

// Next returns a new downlink token.
func (t *DownlinkTokens) Next(correlationIDs []string, time time.Time) uint16 {
	key := uint16(atomic.AddUint32(&t.last, 1))
	pos := key % downlinkTokenItems
	t.items[pos] = downlinkToken{
		key:            key,
		correlationIDs: correlationIDs,
		time:           time,
	}
	return key
}

// Get returns the correlation IDs and time difference between the time given to `Next` and the given time by the token.
// If the token could not be found, this method returns false for `ok`.
func (t DownlinkTokens) Get(token uint16, time time.Time) (correlationIDs []string, delta time.Duration, ok bool) {
	pos := token % downlinkTokenItems
	item := t.items[pos]
	if item.key != token {
		return nil, 0, false
	}
	return item.correlationIDs, time.Sub(item.time), true
}
