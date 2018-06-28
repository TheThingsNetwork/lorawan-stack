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

// Package events implements event handling through a PubSub interface.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Event interface
type Event interface {
	Context() context.Context
	Name() string
	Time() time.Time
	Identifiers() ttnpb.Identifiers
	Data() interface{}
	CorrelationIDs() []string
	Origin() string
	Caller() string
}

func local(evt Event) *event {
	localEvent, ok := evt.(*event)
	if !ok {
		localEvent = &event{innerEvent: innerEvent{
			ctx:            evt.Context(),
			Name:           evt.Name(),
			Time:           evt.Time(),
			Data:           evt.Data(),
			CorrelationIDs: evt.CorrelationIDs(),
			Origin:         evt.Origin(),
			Caller:         evt.Caller(),
		}}
		if ids := evt.Identifiers(); ids != nil {
			localEvent.innerEvent.Identifiers = ids.CombinedIdentifiers()
		}
	}
	return localEvent
}

type event struct {
	innerEvent
}

// IncludeCaller indicates whether the caller of Publish should be included in the event
var IncludeCaller bool

// withCaller returns an event with the Caller field populated, if configured to do so.
// If the original event already had a non-empty Caller, the original event is returned.
func (e *event) withCaller() *event {
	if IncludeCaller && e.innerEvent.Caller == "" {
		if _, file, line, ok := runtime.Caller(2); ok {
			split := strings.SplitAfter(file, "ttn/")
			if len(split) > 1 {
				file = split[1]
			}
			clone := *e
			clone.innerEvent.Caller = fmt.Sprintf("%s:%d", file, line)
			return &clone
		}
	}
	return e
}

type innerEvent struct {
	ctx            context.Context
	Name           string                     `json:"name"`
	Time           time.Time                  `json:"time"`
	Identifiers    *ttnpb.CombinedIdentifiers `json:"identifiers,omitempty"`
	Data           interface{}                `json:"data,omitempty"`
	CorrelationIDs []string                   `json:"correlation_ids,omitempty"`
	Origin         string                     `json:"origin,omitempty"`
	Caller         string                     `json:"caller,omitempty"` // for debugging
}

func (e event) Context() context.Context       { return e.innerEvent.ctx }
func (e event) Name() string                   { return e.innerEvent.Name }
func (e event) Time() time.Time                { return e.innerEvent.Time }
func (e event) Identifiers() ttnpb.Identifiers { return e.innerEvent.Identifiers }
func (e event) Data() interface{}              { return e.innerEvent.Data }
func (e event) CorrelationIDs() []string       { return e.innerEvent.CorrelationIDs }
func (e event) Origin() string                 { return e.innerEvent.Origin }
func (e event) Caller() string                 { return e.innerEvent.Caller }

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// New returns a new Event.
// Event names are dot-separated for namespacing.
// Event data will in most cases be marshaled to JSON, but ideally has (embedded) proto messages.
func New(ctx context.Context, name string, identifiers ttnpb.Identifiers, data interface{}) Event {
	evt := &event{
		innerEvent: innerEvent{
			ctx:            ctx,
			Name:           name,
			Time:           time.Now().UTC(),
			Data:           data,
			Origin:         hostname,
			CorrelationIDs: CorrelationIDsFromContext(ctx),
		},
	}
	if identifiers != nil {
		evt.innerEvent.Identifiers = identifiers.CombinedIdentifiers()
	}
	return evt
}

// UnmarshalJSON unmarshals an event as JSON.
func UnmarshalJSON(data []byte) (Event, error) {
	e := new(event)
	err := json.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}
	e.ctx = context.Background()
	return e, nil
}
