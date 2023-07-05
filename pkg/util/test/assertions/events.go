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

package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

const (
	needEventAssertionCompatible = "This assertion requires a func(events.Event) bool-compatible comparison type (you provided %T)."
	needEventCompatible          = "This assertion requires a Event-compatible comparison type (you provided %T)."
	needEventBuilderCompatible   = "This assertion requires an EventBuilder-compatible comparison type (you provided %T)."
	needEventChannelCompatible   = "This assertion requires a events.Channel-compatible or <-chan test.EventPubSubPublishRequest-compatible comparison type (you provided %T)."
)

// ShouldResembleEvent is used to assert that an events.Event resembles another events.Event.
func ShouldResembleEvent(actual any, expected ...any) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	ee, ok := expected[0].(events.Event)
	if !ok {
		return fmt.Sprintf(needEventCompatible, expected[0])
	}
	ae, ok := actual.(events.Event)
	if !ok {
		return fmt.Sprintf(needEventCompatible, actual)
	}
	ep, err := events.Proto(ee)
	if s := assertions.ShouldBeNil(err); s != success {
		return s
	}
	ap, err := events.Proto(ae)
	if s := assertions.ShouldBeNil(err); s != success {
		return s
	}
	ap.UniqueId = ""
	ep.UniqueId = ""
	ap.Time = nil
	ep.Time = nil
	ap.Authentication = nil
	ep.Authentication = nil
	ap.UserAgent = ""
	ep.UserAgent = ""
	ap.RemoteIp = ""
	ep.RemoteIp = ""
	return ShouldResemble(ap, ep)
}

// ShouldResembleEventBuilder is used to assert that an events.Builder resembles another events.Builder.
func ShouldResembleEventBuilder(actual any, expected ...any) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	ed, ok := expected[0].(events.Builder)
	if !ok {
		return fmt.Sprintf(needEventBuilderCompatible, expected[0])
	}
	ad, ok := actual.(events.Builder)
	if !ok {
		return fmt.Sprintf(needEventBuilderCompatible, actual)
	}
	ctx := context.Background()
	return ShouldResembleEvent(ad.New(ctx), ed.New(ctx))
}

// ShouldResembleEventBuilders is like ShouldResembleEventBuilders, but for events.Builders
func ShouldResembleEventBuilders(actual any, expected ...any) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	eds, ok := expected[0].(events.Builders)
	if !ok {
		return fmt.Sprintf(needEventBuilderCompatible, expected[0])
	}
	ads, ok := actual.(events.Builders)
	if !ok {
		return fmt.Sprintf(needEventBuilderCompatible, actual)
	}
	ctx := context.Background()
	if s := assertions.ShouldHaveLength(ads, len(eds)); s != success {
		return s
	}
	for i, ad := range ads {
		if s := ShouldResembleEvent(ad.New(ctx), eds[i].New(ctx)); s != success {
			return fmt.Sprintf("Mismatch in event definition %d: %s", i, s)
		}
	}
	return success
}

var eventTimeout = test.Delay << 7

func receiveEvent(v any) (events.Event, string) {
	switch ch := v.(type) {
	case <-chan events.Event:
		select {
		case <-time.After(eventTimeout):
			return nil, fmt.Sprintf("Timed out while waiting for event to arrive")
		case ev := <-ch:
			return ev, success
		}
	case <-chan test.EventPubSubPublishRequest:
		select {
		case <-time.After(eventTimeout):
			return nil, fmt.Sprintf("Timed out while waiting for event publish request to arrive")
		case req := <-ch:
			select {
			case <-time.After(eventTimeout):
				return nil, fmt.Sprintf("Timed out while waiting for event publish response to be processed")
			case req.Response <- struct{}{}:
			}
			return req.Event, success
		}
	}
	return nil, fmt.Sprintf(needEventChannelCompatible, v)
}

// ShouldReceiveEventFunc receives 3 parameters. The first being a channel of either events.Event or test.EventPubSubPublishRequest,
// the second being the equality function of type func(events.Event, events.Event) bool and third being the expected events.Event.
func ShouldReceiveEventFunc(actual any, expected ...any) string {
	if len(expected) != 2 {
		return fmt.Sprintf(needExactValues, 2, len(expected))
	}
	eq, ok := expected[0].(func(events.Event, events.Event) bool)
	if !ok {
		return fmt.Sprintf(needEventAssertionCompatible, expected[0])
	}
	ee, ok := expected[1].(events.Event)
	if !ok {
		return fmt.Sprintf(needEventCompatible, expected[1])
	}
	ae, s := receiveEvent(actual)
	if s != success {
		return s
	}
	return assertions.ShouldBeTrue(eq(ae, ee))
}

// ShouldReceiveEventResembling is like ShouldReceiveEventFunc, but uses same resemblance function as ShouldResembleEvent.
func ShouldReceiveEventResembling(actual any, expected ...any) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	ae, s := receiveEvent(actual)
	if s != success {
		return s
	}
	return ShouldResembleEvent(ae, expected[0])
}

func eventSlice(vs ...any) ([]events.Event, string) {
	var evs []events.Event
	for _, v := range vs {
		ev, ok := v.(events.Event)
		if ok {
			evs = append(evs, ev)
			continue
		}
		r, ok := test.WrapRanger(v)
		if !ok {
			return nil, fmt.Sprintf("Cannot range over values of type %T", v)
		}
		s := success
		r.Range(func(_, v any) bool {
			ev, ok := v.(events.Event)
			if !ok {
				s = fmt.Sprintf(needEventCompatible, v)
				return false
			}
			evs = append(evs, ev)
			return true
		})
		if s != success {
			return nil, s
		}
	}
	return evs, success
}

// ShouldReceiveEventsFunc is like ShouldReceiveEventFunc, but allows for several expected events to be specified.
// Expected events should be passed as variadic parameters, which can be wrapped any collection of events.Event, that test.WrapRanger can range over.
func ShouldReceiveEventsFunc(actual any, expected ...any) string {
	if len(expected) < 2 {
		return fmt.Sprintf(needAtLeastValues, 2, len(expected))
	}
	evs, s := eventSlice(expected[1:]...)
	if s != success {
		return s
	}
	for i, exp := range evs {
		if s := ShouldReceiveEventFunc(actual, expected[0], exp); s != success {
			return fmt.Sprintf("Mismatch in event number %d: %s", i, s)
		}
	}
	return success
}

// ShouldReceiveEventsResembling is like ShouldReceiveEventsFunc, but uses same resemblance function as ShouldResembleEvent.
func ShouldReceiveEventsResembling(actual any, expected ...any) string {
	if len(expected) == 0 {
		return fmt.Sprintf(needAtLeastValues, 1, len(expected))
	}
	evs, s := eventSlice(expected...)
	if s != success {
		return s
	}
	for i, exp := range evs {
		if s := ShouldReceiveEventResembling(actual, exp); s != success {
			return fmt.Sprintf("Mismatch in event number %d:\n%s", i, s)
		}
	}
	return success
}
