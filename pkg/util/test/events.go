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

package test

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/pkg/events"
)

// EventsChannel used for testing event delivery.
type EventsChannel events.Channel

// CollectEvents returns a channel where events matching name will be added.
func CollectEvents(name string) EventsChannel {
	collectedEvents := make(events.Channel, 32)
	events.Subscribe(name, collectedEvents)
	return EventsChannel(collectedEvents)
}

// Expect n events, fail the test if not received within reasonable time.
func (ch EventsChannel) Expect(t *testing.T, n int) []events.Event {
	events := make([]events.Event, 0, n)
	for i := 0; i < n; i++ {
		select {
		case evt := <-ch:
			events = append(events, evt)
		case <-time.After(10 * time.Millisecond * Delay):
			t.Fatalf("Did not receive expected event %d/%d", i+1, n)
		}
	}
	return events
}
