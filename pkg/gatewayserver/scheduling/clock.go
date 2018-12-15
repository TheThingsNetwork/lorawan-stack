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

package scheduling

import (
	"math"
	"time"
)

// Clock represents an absolute time source.
type Clock interface {
	// Now returns an indication of the current concentrator time at the given server time.
	Now(server time.Time) time.Duration
	// ConcentratorTime returns the concentrator time based on the given absolute time.
	ConcentratorTime(time.Time) time.Duration
}

// RolloverClock is a Clock that takes roll-over of uint32 concentrator time into account.
type RolloverClock struct {
	relative        uint32
	absolute        time.Duration
	server, gateway time.Time
}

// Sync synchronizes the clock with the given concentrator time v, the server time and the gateway time that
// corresponds to the given v.
func (c *RolloverClock) Sync(v uint32, server, gateway time.Time) {
	passed := int64(v) - int64(c.relative)
	if passed < 0 {
		passed += math.MaxUint32
	}
	c.relative = v
	c.absolute = c.absolute + time.Duration(passed)*time.Microsecond
	c.server = server
	c.gateway = gateway
}

// Now implements Clock.
func (c *RolloverClock) Now(server time.Time) time.Duration {
	return c.absolute + server.Sub(c.server)
}

// ConcentratorTime implements Clock.
func (c *RolloverClock) ConcentratorTime(t time.Time) time.Duration {
	return c.absolute + t.Sub(c.gateway)
}
